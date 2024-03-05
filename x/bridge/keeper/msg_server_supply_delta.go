package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (k msgServer) SupplyDelta(goCtx context.Context, msg *types.MsgSupplyDelta) (*types.MsgSupplyDeltaResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Confirm signer is the module

	// TODO: Confirm that supply_delta_period is non-zero

	// TODO: Check if height from ctx % params.supply_delta_period is 0 to confirm that it was injected correctly.

	// TODO: Handling the message and construct full response
	_ = ctx

	// TODO: Refactor to get and increment (see AddRaw postfixed)
	nonce := k.MustGetLastEthereumNonce(ctx).AddRaw(1)
	k.SetLastEthereumNonce(ctx, nonce)

	// TODO: Implement this logic
	//     deltaToReport = delta + offset

	// TODO: Reject message if deltaToReport is zero (just in case user submits msgSupplyDelta at the same height of
	//     : required height to be injected.

	// TODO: Set offset and delta to zero

	// TODO: Add event, we can have one event that reports supply

	return &types.MsgSupplyDeltaResponse{
		Nonce: nonce,
		// SupplyDelta: Put negative or positive depending on mint or burn
	}, nil
}
