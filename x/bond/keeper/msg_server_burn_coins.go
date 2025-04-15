package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/metrics"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

// BurnCoins implements the Msg/BurnCoins RPC method.
func (k msgServer) BurnCoins(goCtx context.Context, msg *types.MsgBurnCoins) (*types.MsgBurnCoinsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Coins.Empty() {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidCoins, "coins cannot be empty")
	}

	sender, err := k.GetAddressCodec().StringToBytes(msg.Sender)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "failed to decode from address")
	}

	// Send coins from sender to module account
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, msg.Coins); err != nil {
		return nil, err
	}

	// Burn the coins
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, msg.Coins); err != nil {
		return nil, err
	}

	// Record metrics
	metrics.ObserveBurnCoins(goCtx, msg.Coins)

	return &types.MsgBurnCoinsResponse{}, nil
}
