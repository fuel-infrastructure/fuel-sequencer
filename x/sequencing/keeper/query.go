package keeper

import (
	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"
)

var _ types.QueryServer = Keeper{}
