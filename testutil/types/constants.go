package types

import (
	sdk "cosmossdk.io/math"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	_ "github.com/fuel-infrastructure/fuel-sequencer/app/apptesting" // Required to load the right config for testing
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

var (
	TestGovernanceAddress = authtypes.NewModuleAddress(govtypes.ModuleName).String()
	TestSupplyDeltaPeriod = uint64(100)
	TestLastEthereumNonce = sdk.NewInt(50)
	TestLastSupply        = sdk.NewInt(100000000)
	TestDelta             = sdk.NewInt(5000000)
	TestOffset            = sdk.NewInt(-2000000)
	TestSupplyDeltaInfo   = bridgetypes.SupplyDeltaInfo{
		LastSupply: TestLastSupply,
		Delta:      TestDelta,
		Offset:     TestOffset,
	}
)
