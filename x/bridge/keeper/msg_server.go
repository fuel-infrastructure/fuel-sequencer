package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/utils"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
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

// TryExecSpecialMessage runs a special message and records a failure in the index if this message fails. This function
// only returns an error if the signer is not authorised to execute the message, i.e. is not the authority address.
func (k msgServer) TryExecSpecialMessage(ctx sdk.Context, signer string, msg func(ctx sdk.Context) error) error {

	// Confirm that the msg signer is the bridge module's authority address (governance).
	if k.GetAuthority() != signer {
		return types.ErrInvalidSigner.Wrapf(
			"invalid authority; expected %s, got %s", k.GetAuthority(), signer,
		)
	}

	// Catch failures of special messages`
	err := utils.ApplyFuncIfNoErrorAndNoPanic(ctx, msg)
	if err != nil {
		index, found := k.GetIndex(ctx)
		if !found {
			index = types.Index{NumFailedSpecialTxs: 0}
		}
		index.NumFailedSpecialTxs += 1
	}

	return nil
}
