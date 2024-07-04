package testsuite

import (
	"fmt"
	"time"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	libclient "github.com/cometbft/cometbft/rpc/jsonrpc/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	grouptypes "github.com/cosmos/cosmos-sdk/x/group"
	paramsproposaltypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/ethclient"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	commitmentstypes "github.com/fuel-infrastructure/fuel-sequencer/x/commitments/types"
	minttypes "github.com/fuel-infrastructure/fuel-sequencer/x/mint/types"
	sequencingtypes "github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCClients holds a reference to any GRPC clients that are needed by the tests.
// These should typically be used for query clients only. If we need to make changes, we should
// use E2ETestSuite.BroadcastMessages to broadcast transactions instead.
type GRPCClients struct {

	// Cosmos SDK query clients
	GovQueryClient          govtypesv1.QueryClient
	GroupsQueryClient       grouptypes.QueryClient
	ParamsQueryClient       paramsproposaltypes.QueryClient
	AuthQueryClient         authtypes.QueryClient
	AuthZQueryClient        authz.QueryClient
	BankQueryClient         banktypes.QueryClient
	DistributionQueryClient distributiontypes.QueryClient
	ConsensusQueryClient    consensustypes.QueryClient
	StakingQueryClient      stakingtypes.QueryClient
	MintQueryClient         minttypes.QueryClient

	// Custom query clients
	BridgeQueryClient      bridgetypes.QueryClient
	SequencingQueryClient  sequencingtypes.QueryClient
	CommitmentsQueryClient commitmentstypes.QueryClient

	ConsensusServiceClient cmtservice.ServiceClient
}

// initGRPCClients establishes GRPC clients using the first validator.
func (s *E2ETestSuite) initGRPCClients() {
	addr := s.Chain.validators[0].hostGRPCPort

	// Create a connection to the gRPC server.
	grpcConn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	s.Require().NoError(err)
	s.T().Cleanup(func() {
		if err := grpcConn.Close(); err != nil {
			s.T().Logf("failed closing GRPC connection to Chain %s: %s", s.Chain.id, err)
		}
	})

	s.Chain.grpcClients = &GRPCClients{
		GovQueryClient:          govtypesv1.NewQueryClient(grpcConn),
		GroupsQueryClient:       grouptypes.NewQueryClient(grpcConn),
		ParamsQueryClient:       paramsproposaltypes.NewQueryClient(grpcConn),
		AuthQueryClient:         authtypes.NewQueryClient(grpcConn),
		AuthZQueryClient:        authz.NewQueryClient(grpcConn),
		BankQueryClient:         banktypes.NewQueryClient(grpcConn),
		DistributionQueryClient: distributiontypes.NewQueryClient(grpcConn),
		ConsensusQueryClient:    consensustypes.NewQueryClient(grpcConn),
		BridgeQueryClient:       bridgetypes.NewQueryClient(grpcConn),
		SequencingQueryClient:   sequencingtypes.NewQueryClient(grpcConn),
		CommitmentsQueryClient:  commitmentstypes.NewQueryClient(grpcConn),
		ConsensusServiceClient:  cmtservice.NewServiceClient(grpcConn),
		StakingQueryClient:      stakingtypes.NewQueryClient(grpcConn),
		MintQueryClient:         minttypes.NewQueryClient(grpcConn),
	}
}

func (s *E2ETestSuite) getGRPCClients() *GRPCClients {
	return s.Chain.grpcClients
}

// initRPCClient establishes an RPC client using the first validator.
func (s *E2ETestSuite) initRPCClient() {
	addr := s.Chain.validators[0].hostRPCPort

	httpClient, err := libclient.DefaultHTTPClient(addr)
	if err != nil {
		panic(err)
	}

	httpClient.Timeout = 10 * time.Second
	rpcClient, err := rpchttp.NewWithClient(addr, "/websocket", httpClient)
	if err != nil {
		panic(err)
	}

	s.Chain.rpcClient = rpcClient
}

func (s *E2ETestSuite) getRPCClient() *rpchttp.HTTP {
	return s.Chain.rpcClient
}

// initEthereumRPCClient establishes an RPC client to the Ethereum node.
func (s *E2ETestSuite) initEthereumRPCClient() {

	url := fmt.Sprintf("http://%s", s.ethResource.GetHostPort("8545/tcp"))
	ethClient, err := ethclient.Dial(url)
	s.Require().NoError(err)

	s.Chain.ethClient = ethClient
}

func (s *E2ETestSuite) getEthereumRPCClient() *ethclient.Client {
	return s.Chain.ethClient
}

// initSidecarClient establishes a Sidecar client using the first validator.
func (s *E2ETestSuite) initSidecarClient() {
	addr := s.Chain.validators[0].sidecarGRPCPort

	// Create a connection to the gRPC server.
	grpcConn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	s.Require().NoError(err)
	s.T().Cleanup(func() {
		if err := grpcConn.Close(); err != nil {
			s.T().Logf("failed closing GRPC connection to sidecar: %s", err)
		}
	})

	s.Chain.sidecarClient = sidecartypes.NewSidecarClient(grpcConn)
}

func (s *E2ETestSuite) getSidecarClient() sidecartypes.SidecarClient {
	return s.Chain.sidecarClient
}
