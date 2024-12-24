package keeper

import (
	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"
)

var _ types.QueryServer = Keeper{}
