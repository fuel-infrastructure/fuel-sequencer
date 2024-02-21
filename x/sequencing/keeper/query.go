package keeper

import (
	"fuelsequencer/x/sequencing/types"
)

var _ types.QueryServer = Keeper{}
