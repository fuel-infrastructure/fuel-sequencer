package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (k msgServer) DepositFromEthereum(goCtx context.Context, msg *types.MsgDepositFromEthereum) (*types.MsgDepositFromEthereumResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	params := k.GetParams(ctx)
	supplyDeltaInfo := k.MustGetSupplyDeltaInfo(ctx)

	k.processDepositEvent(ctx, msg, &params, &supplyDeltaInfo)

	return &types.MsgDepositFromEthereumResponse{}, nil
}
