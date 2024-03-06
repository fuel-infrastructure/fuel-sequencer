package types

import (
	sdkmath "cosmossdk.io/math"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

var (
	TestGovernanceAddress = authtypes.NewModuleAddress(govtypes.ModuleName).String()
	TestSupplyDeltaPeriod = uint64(100)
	TestLastEthereumNonce = sdkmath.NewInt(50)
	TestLastSupply        = sdkmath.NewInt(100000000)
	TestDelta             = sdkmath.NewInt(5000000)
	TestOffset            = sdkmath.NewInt(-2000000)
	TestSupplyDeltaInfo   = bridgetypes.SupplyDeltaInfo{
		LastSupply: TestLastSupply,
		Delta:      TestDelta,
		Offset:     TestOffset,
	}
)
