package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (k msgServer) WithdrawToEthereum(
	goCtx context.Context,
	msg *types.MsgWithdrawToEthereum,
) (*types.MsgWithdrawToEthereumResponse, error) {

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Retrieve bridge parameters
	params := k.GetParams(ctx)
	bridgeDenom := params.BridgeDenom

	// Validate that the token is the bridge token
	if msg.Amount.Denom != bridgeDenom {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"invalid token denom: %s, expected: %s",
			msg.Amount.Denom,
			bridgeDenom,
		)
	}

	// Burn the user's bridge tokens
	withdrawerAccAddress, err := k.GetAddressCodec().StringToBytes(msg.From)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "failed to decode from address")
	}
	if err := k.BurnCoinsFromAddress(ctx, withdrawerAccAddress, sdk.NewCoins(msg.Amount)); err != nil {
		return nil, errorsmod.Wrapf(err, "failed to burn bridge tokens")
	}

	// Apply the offset to the supplyDelta
	supplyDeltaInfo := k.MustGetSupplyDeltaInfo(ctx)
	supplyDeltaInfo.Offset = supplyDeltaInfo.Offset.Add(msg.Amount.Amount)
	k.SetSupplyDeltaInfo(ctx, supplyDeltaInfo)

	// Update the Ethereum nonce
	nonce := k.MustGetLastEthereumNonce(ctx).AddRaw(1)
	k.SetLastEthereumNonce(ctx, nonce)

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(
		&types.EventWithdrawToEthereumReported{
			Nonce:  nonce,
			From:   msg.From,
			To:     msg.To,
			Amount: msg.Amount,
		},
	)
	if err != nil {
		return nil, err
	}

	return &types.MsgWithdrawToEthereumResponse{
		Nonce:  nonce,
		From:   msg.From,
		To:     msg.To,
		Amount: msg.Amount,
	}, nil
}
