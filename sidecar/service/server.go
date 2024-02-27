package sidecar

import (
	"context"
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
	"google.golang.org/grpc/credentials/insecure"

	"github.com/fuel-infrastructure/fuel-sequencer/sidecar"
	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/sync"
)

const DefaultServerShutdownTimeout = 3 * time.Second

// SidecarServer is the base implementation of the service.SidecarServer interface, this is meant to
// serve requests from a remote SidecarClient.
type SidecarServer struct { //nolint
	types.UnimplementedSidecarServer

	// expected implementation of the sidecar
	o sidecar.Sidecar

	// underlying grpc-server -- serves all grpc requests
	grpcSrv *grpc.Server

	// grpc-gateway mux -- serves all http grpc proxy requests
	gatewayMux *runtime.ServeMux

	// underlying http server
	httpSrv *http.Server

	// closer to handle graceful closures from multiple go-routines
	*sync.Closer

	// logger to log incoming requests
	logger *zap.Logger
}

// NewSidecarServer returns a new instance of the SidecarServer, given an implementation of the Sidecar interface.
func NewSidecarServer(o sidecar.Sidecar, logger *zap.Logger) *SidecarServer {
	logger = logger.With(zap.String("server", "sidecar"))

	os := &SidecarServer{
		o:      o,
		logger: logger,
	}
	os.Closer = sync.NewCloser().WithCallback(func() {
		// if the server has been started, close it
		if os.httpSrv != nil {
			ctx, cf := context.WithTimeout(context.Background(), DefaultServerShutdownTimeout)
			os.httpSrv.Shutdown(ctx) // close HTTP server backing GRPC-gateway
			os.grpcSrv.Stop()        // close GRPC server serving listeners that have been routed to GRPC server
			cf()
		}
	})

	return os
}

// routeRequest determines if the incoming http request is a grpc or http request and routes to the proper handler.
func (os *SidecarServer) routeRequest(w http.ResponseWriter, r *http.Request) {
	if r.ProtoMajor == 2 && strings.HasPrefix(
		r.Header.Get("Content-Type"), "application/grpc") {

		os.grpcSrv.ServeHTTP(w, r)
	} else {
		os.gatewayMux.ServeHTTP(w, r)
	}
}

// StartServer starts the sidecar gRPC server on the given host and port. The server is killed on any errors from the listener, or if ctx is cancelled.
// This method returns an error via any failure from the listener. This is a blocking call, i.e until the server is closed or the server errors,
// this method will block.
func (os *SidecarServer) StartServer(ctx context.Context, host, port string) error {
	serverEndpoint := fmt.Sprintf("%s:%s", host, port)
	os.httpSrv = &http.Server{
		Addr:              serverEndpoint,
		ReadHeaderTimeout: DefaultServerShutdownTimeout,
	}
	// create grpc server
	os.grpcSrv = grpc.NewServer()
	// register sidecar server
	types.RegisterSidecarServer(os.grpcSrv, os)

	// register the grpc-gateway
	// it handles the http request and dials the server endpoint with the grpc request
	os.gatewayMux = runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &gateway.JSONPb{
			EmitDefaults: true,
			Indent:       "",
			OrigName:     true,
		}),
	)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := types.RegisterSidecarHandlerFromEndpoint(ctx, os.gatewayMux, serverEndpoint, opts)
	if err != nil {
		return err
	}

	router := http.NewServeMux()
	router.HandleFunc("/", os.routeRequest)

	os.httpSrv.Handler = h2c.NewHandler(router, &http2.Server{})

	eg, ctx := errgroup.WithContext(ctx)

	// listen for ctx cancellation
	eg.Go(func() error {
		// if the context is closed, close the server + sidecar
		<-ctx.Done()
		os.logger.Info("context cancelled, closing sidecar")

		_ = os.Close()
		return nil
	})

	// start the sidecar, return error if it fails
	eg.Go(func() error {
		os.logger.Info("starting sidecar")
		return os.o.Start(ctx)
	})

	// start the server
	eg.Go(func() error {
		// serve, and return any errors
		os.logger.Info(
			"starting grpc server",
			zap.String("host", host),
			zap.String("port", port),
		)

		err = os.httpSrv.ListenAndServe()
		if err != nil {
			return fmt.Errorf("[grpc server]: error serving: %w", err)
		}

		return nil
	})

	// wait for everything to finish
	return eg.Wait()
}

func (os *SidecarServer) GetBlockEvents(
	ctx context.Context,
	req *types.QueryBlockEventsRequest,
) (*types.QueryBlockEventsResponse, error) {
	// Check that the request is non-nil
	if req == nil {
		return nil, errors.New("nil request")
	}

	os.logger.Info("received request for block events", zap.String("blockNumber", req.BlockNumber))

	// Check that sidecar is running
	if !os.o.IsRunning() {
		os.logger.Error("sidecar not running")
		return nil, errors.New("sidecar not running")
	}

	// Convert the block number from the request to a big.Int
	blockNumber, ok := new(big.Int).SetString(req.BlockNumber, 10)
	if !ok {
		return nil, errors.New("invalid block number")
	}

	resCh := make(chan *types.QueryBlockEventsResponse)

	// Run the request in a goroutine, to unblock server + ctx cancellation
	go func() {
		var events []*types.Event

		blockchainEvents, err := os.o.QueryBlockEvents(blockNumber)
		if err != nil {
			os.logger.Error("error querying block events", zap.Error(err))
			resCh <- nil
			return
		}

		// Convert blockchain events to protobuf `Event` type
		for _, be := range blockchainEvents {
			events = append(events, &types.Event{
				EventType: be.EventType,
				Data:      be.Data,
			})
		}

		resCh <- &types.QueryBlockEventsResponse{Events: events}
	}()

	// Defer to context closure
	select {
	case <-ctx.Done():
		os.logger.Error("context cancelled")
		return nil, context.Canceled
	case resp := <-resCh:
		if resp == nil {
			return nil, errors.New("failed to get block events")
		}
		return resp, nil
	}
}

// Close closes the underlying sidecar server, and blocks until all open requests have been satisfied.
func (os *SidecarServer) Close() error {
	// close + close server if necessary
	os.Closer.Close()
	return nil
}

// Done returns a channel that is closed when the sidecar server is closed.
func (os *SidecarServer) Done() <-chan struct{} {
	return os.Closer.Done()
}
