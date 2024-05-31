package service

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"github.com/fuel-infrastructure/fuel-sequencer/x/commitments/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

func RegisterCommitmentsService(
	clientCtx client.Context,
	server gogogrpc.Server,
	iRegistry codectypes.InterfaceRegistry,
) {
	types.RegisterQueryServer(server, NewQueryServer(clientCtx, iRegistry))
}

func RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}
