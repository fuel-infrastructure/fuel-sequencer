package keeper

import (
	"github.com/fuel-infrastructure/fuel-sequencer/x/blob/types"
)

var _ types.QueryServer = (*Keeper)(nil)
