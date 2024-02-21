package keeper

import (
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

var _ types.QueryServer = Keeper{}
