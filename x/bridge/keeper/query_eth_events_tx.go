package keeper

import (
	"context"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) EthEventsTxByBlockNumber(goCtx context.Context, req *types.QueryGetEthEventsTxByBlockNumberRequest) (*types.QueryGetEthEventsTxByBlockNumberResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// Convert block_number from string to uint64
	blockNumber, err := strconv.ParseUint(req.BlockNumber, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "could not convert block number to uint64: %v", err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Use the GetEthEventsTx method to fetch the EthEventsTx data
	val, found := k.GetEthEventsTx(ctx, blockNumber)
	if !found {
		return nil, status.Error(codes.NotFound, "EthEventsTx not found for block number")
	}

	// Prepare and return the response
	return &types.QueryGetEthEventsTxByBlockNumberResponse{
		EthEventsTx: &val,
	}, nil
}
