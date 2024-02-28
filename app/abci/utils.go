package abci

import sdk "github.com/cosmos/cosmos-sdk/types"

// VoteExtensionsEnabled determines if vote extensions are enabled for the current block. If vote extensions are enabled
// at height h, then a proposer will receive vote extensions in height h+1. This is primarily utilized by any module
// that needs to make state changes based on whether vote extensions have been included in a proposal.
func VoteExtensionsEnabled(ctx sdk.Context) bool {
	cp := ctx.ConsensusParams()
	if cp.Abci == nil || cp.Abci.VoteExtensionsEnableHeight == 0 {
		return false
	}

	return cp.Abci.VoteExtensionsEnableHeight < ctx.BlockHeight()
}
