package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (k Keeper) SequencerAddressFromEthereumAddress(
	_ context.Context,
	req *types.QuerySequencerAddressFromEthereumAddressRequest,
) (*types.QuerySequencerAddressFromEthereumAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	sequencerAddress, err := k.GenerateSequencerAddressFromEthereumAddress(req.EthereumAddress)
	if err != nil {
		return nil, err
	}

	return &types.QuerySequencerAddressFromEthereumAddressResponse{
		SequencerAddress: sequencerAddress.String(),
	}, nil
}
