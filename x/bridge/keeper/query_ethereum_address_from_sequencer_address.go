package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (k Keeper) EthereumAddressFromSequencerAddress(_ context.Context, req *types.QueryEthereumAddressFromSequencerAddressRequest) (*types.QueryEthereumAddressFromSequencerAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ethereumAddress, err := k.GenerateEthereumAddressFromSequencerAddress(req.SequencerAddress)
	if err != nil {
		return nil, err
	}

	return &types.QueryEthereumAddressFromSequencerAddressResponse{
		EthereumAddress: ethereumAddress.String(),
	}, nil
}
