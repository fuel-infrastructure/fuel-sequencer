package keeper

import (
	"fuelsequencer/x/bridge/types"
)

var _ types.QueryServer = Keeper{}
