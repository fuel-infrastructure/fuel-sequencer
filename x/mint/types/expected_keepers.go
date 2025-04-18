package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	bondtypes "github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// BridgeKeeper defines the contract needed to be fulfilled for bridge module dependencies.
type BridgeKeeper interface {
	GetParams(ctx context.Context) bridgetypes.Params
}

// BondKeeper defines the contract needed to be fulfilled for bond module dependencies.
type BondKeeper interface {
	GetParams(ctx context.Context) bondtypes.Params
	AddCollectedBondStake(ctx context.Context, bond sdk.Coins) error
}
