package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (k msgServer) DepositFromEthereum(
	goCtx context.Context, msg *types.MsgDepositFromEthereum,
) (resp *types.MsgDepositFromEthereumResponse, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err = k.TryExecSpecialMessage(ctx, msg.Authority, func(ctx sdk.Context) error {
		var innerErr error
		if resp, innerErr = k.depositFromEthereum(ctx, msg); innerErr != nil {
			resp = &types.MsgDepositFromEthereumResponse{} // normalise
			return innerErr
		}
		return nil
	})

	return resp, err
}

func (k msgServer) depositFromEthereum(
	ctx sdk.Context, msg *types.MsgDepositFromEthereum,
) (*types.MsgDepositFromEthereumResponse, error) {

	params := k.GetParams(ctx)
	supplyDeltaInfo := k.MustGetSupplyDeltaInfo(ctx)

	k.processDepositEvent(ctx, msg, &params, &supplyDeltaInfo)

	return &types.MsgDepositFromEthereumResponse{}, nil
}
