package service

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	gateway "github.com/cosmos/gogogateway"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/sidecar"
	"github.com/fuel-infrastructure/fuel-sequencer/utils/credentials"
)

const DefaultServerShutdownTimeout = 3 * time.Second

// queryBlockEventsResponseWithError struct to hold both a response and an error
type queryBlockEventsResponseWithError struct {
	Response *types.QueryBlockEventsResponse
	Err      error
}

// SidecarServer is the base implementation of the service.SidecarServer interface, this is meant to
// serve requests from a remote SidecarClient.
type SidecarServer struct { //nolint
	types.UnimplementedSidecarServer

	// expected implementation of the sidecar
	s sidecar.SidecarI

	// underlying grpc-server -- serves all grpc requests
	grpcSrv *grpc.Server

	// grpc-gateway mux -- serves all http grpc proxy requests
	gatewayMux *runtime.ServeMux

	// underlying http server
	httpSrv *http.Server

	// logger to log incoming requests
	logger *zap.Logger

	// shutdownCh to process any shutdown signals
	shutdownCh chan struct{}

	// metrics is the set of all Prometheus metrics exposed by SidecarServer.
	metrics *Metrics
}

// NewSidecarServer returns a new instance of the SidecarServer, given an implementation of the Sidecar interface.
func NewSidecarServer(s sidecar.SidecarI, logger *zap.Logger, metrics *Metrics) *SidecarServer {
	ss := &SidecarServer{
		s:          s,
		logger:     logger,
		shutdownCh: make(chan struct{}),
		metrics:    metrics,
	}

	return ss
}

func (ss *SidecarServer) InitializeServer(host, port, pathToCertFile, pathToKeyFile string) error {
	serverEndpoint := fmt.Sprintf("%s:%s", host, port)

	// Make use of certificates if indicated by the operator
	serverCredentials, certificates, tlsEnabled, err := credentials.NewServerTransportCredentialsFromCertFile(
		pathToCertFile, pathToKeyFile,
	)
	if err != nil {
		panic(fmt.Errorf("failed to get sidecar server TLS credentials; error: %w", err))
	}

	ss.grpcSrv = grpc.NewServer(grpc.Creds(serverCredentials))
	types.RegisterSidecarServer(ss.grpcSrv, ss)

	ss.gatewayMux = runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &gateway.JSONPb{
			EmitDefaults: true,
			Indent:       "",
			OrigName:     true,
		}),
	)

	opts := []grpc.DialOption{grpc.WithTransportCredentials(serverCredentials)}
	if err = types.RegisterSidecarHandlerFromEndpoint(
		context.Background(), ss.gatewayMux, serverEndpoint, opts,
	); err != nil {
		return err
	}

	router := http.NewServeMux()
	router.HandleFunc("/", ss.routeRequest)

	ss.httpSrv = &http.Server{
		Addr:              serverEndpoint,
		Handler:           h2c.NewHandler(router, &http2.Server{}),
		ReadHeaderTimeout: DefaultServerShutdownTimeout,
	}

	if tlsEnabled {
		ss.httpSrv.TLSConfig = &tls.Config{
			MinVersion:   tls.VersionTLS12,
			Certificates: certificates,
		}
	}

	return nil
}

func (ss *SidecarServer) StartServer(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	// Listen for context cancellation to stop server.
	eg.Go(func() error {
		<-ctx.Done()
		ss.logger.Info("context cancelled, closing sidecar")
		return ss.ShutdownServer()
	})

	// Start processing events from the sidecar.
	eg.Go(func() error {
		return ss.s.Start(ctx)
	})

	eg.Go(func() error {
		ss.logger.Info("starting grpc server", zap.String("address", ss.httpSrv.Addr))
		if ss.httpSrv.TLSConfig != nil {
			// Use ListenAndServeTLS for HTTPS. certFile and keyFile can be left as empty string as we expect the
			// certificate to be loaded in the initializeServer function.
			if err := ss.httpSrv.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
				return fmt.Errorf("[grpc server] server ListenAndServeTLS: %w", err)
			}
		} else {
			// Use ListenAndServe for HTTP
			if err := ss.httpSrv.ListenAndServe(); err != http.ErrServerClosed {
				return fmt.Errorf("[grpc server] server ListenAndServe: %w", err)
			}
		}
		return nil
	})

	return eg.Wait()
}

func (ss *SidecarServer) ShutdownServer() error {
	ss.logger.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), DefaultServerShutdownTimeout)
	defer cancel()

	if err := ss.httpSrv.Shutdown(ctx); err != nil {
		ss.logger.Error("error during server shutdown", zap.Error(err))
		return err
	}

	ss.grpcSrv.GracefulStop()
	ss.logger.Info("server shutdown completed")
	return nil
}

func (ss *SidecarServer) GetBlockEvents(
	ctx context.Context,
	req *types.QueryBlockEventsRequest,
) (*types.QueryBlockEventsResponse, error) {
	// Check that the request is non-nil
	if req == nil {
		return nil, errors.New("nil request")
	}

	ss.logger.Info("received request for events", zap.String("block", req.BlockNumber))

	// Check that sidecar is running
	if ss.s.IsStopped() {
		ss.logger.Error("sidecar not running")
		return nil, errors.New("sidecar not running")
	}

	// Convert the block number from the request to a big.Int
	blockNumber, ok := new(big.Int).SetString(req.BlockNumber, 10)
	if !ok {
		return nil, errors.New("invalid block number")
	}

	resCh := make(chan *queryBlockEventsResponseWithError)

	// Run the request in a goroutine, to unblock server + ctx cancellation
	go func() {
		var events []*types.Event

		blockchainEvents, err := ss.s.QueryBlockEvents(blockNumber)
		if err != nil {
			ss.logger.Warn("error processing events query", zap.Error(err))
			defer ss.metrics.BlockEventsRequestsTotal.With("success", "false").Add(1)
			resCh <- &queryBlockEventsResponseWithError{Response: nil, Err: err}
			return
		}

		// Use pointer for blockchain events
		// TODO: Consider not having to convert to a pointer...
		for _, be := range blockchainEvents {
			events = append(events, &be)
		}

		defer ss.metrics.BlockEventsRequestsTotal.With("success", "true").Add(1)
		resCh <- &queryBlockEventsResponseWithError{Response: &types.QueryBlockEventsResponse{Events: events}, Err: nil}
	}()

	// Defer to context closure
	select {
	case <-ctx.Done():
		ss.logger.Error("context cancelled")
		return nil, context.Canceled
	case resp := <-resCh:
		if resp.Err != nil {
			return nil, resp.Err
		}
		return resp.Response, nil
	}
}

// routeRequest determines if the incoming http request is a grpc or http request and routes to the proper handler.
func (ss *SidecarServer) routeRequest(w http.ResponseWriter, r *http.Request) {
	if r.ProtoMajor == 2 && strings.HasPrefix(
		r.Header.Get("Content-Type"), "application/grpc") {

		ss.grpcSrv.ServeHTTP(w, r)
	} else {
		ss.gatewayMux.ServeHTTP(w, r)
	}
}
