package service

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/fuel-infrastructure/fuel-sequencer/x/commitments/types"
)

// RegisterCommitmentsService registers the bridge commitments queries on the gRPC router.
func RegisterCommitmentsService(
	clientCtx client.Context,
	server gogogrpc.Server,
	iRegistry codectypes.InterfaceRegistry,
	maxQueryRange uint64,
) {
	types.RegisterQueryServer(server, NewQueryServer(clientCtx, iRegistry, maxQueryRange))
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for bridge commitments.
func RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}
