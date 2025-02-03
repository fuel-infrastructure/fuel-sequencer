package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// BridgeKeeper defines the contract needed to be fulfilled for bridge module dependencies.
type BridgeKeeper interface {
	GetParams(ctx context.Context) bridgetypes.Params
}

// MintBankKeeper defines the contract needed to be fulfilled for the bank module dependencies for minting purposes.
type MintBankKeeper interface {
	MintCoins(ctx context.Context, moduleName string, coins sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
}
