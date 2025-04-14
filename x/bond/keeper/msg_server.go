package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// BurnCoins implements the Msg/BurnCoins RPC method.
func (k msgServer) BurnCoins(goCtx context.Context, msg *types.MsgBurnCoins) (*types.MsgBurnCoinsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	// Send coins from sender to module account
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, msg.Coins); err != nil {
		return nil, err
	}

	// Burn the coins
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, msg.Coins); err != nil {
		return nil, err
	}

	return &types.MsgBurnCoinsResponse{}, nil
}
