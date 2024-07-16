package types

import (
	"context"

	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// BridgeKeeper defines the contract needed to be fulfilled for bridge module dependencies.
type BridgeKeeper interface {
	GetParams(ctx context.Context) bridgetypes.Params
}
