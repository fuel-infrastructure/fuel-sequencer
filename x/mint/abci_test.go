package mint_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/mint"
)

func (s *MintModuleTestSuite) TestBeginBlocker_InflationBasedOnBridgeModuleParams() {

	bondDenom := sdk.DefaultBondDenom
	inflation := sdkmath.LegacyMustNewDecFromStr("0.1")
	inflationCalculationFn := minttypes.DefaultInflationCalculationFn
	feeCollector := s.App.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)

	// Simplify mint module params so that we have a constant 10% inflation.
	mintParams, err := s.App.MintKeeper.Params.Get(s.Ctx())
	s.Require().NoError(err)
	mintParams.InflationMin = inflation
	mintParams.InflationMax = inflation
	mintParams.InflationRateChange = sdkmath.LegacyMustNewDecFromStr("0.0")
	s.Require().NoError(s.App.MintKeeper.Params.Set(s.Ctx(), mintParams))

	// Override BridgeDenomTotalSupply so that we know what supply value will be used.
	bridgeParams := s.App.BridgeKeeper.GetParams(s.Ctx())
	bridgeParams.BridgeDenom = bondDenom
	bridgeParams.BridgeDenomTotalSupply = sdkmath.NewInt(10_000_000_000)
	s.Require().NoError(s.App.BridgeKeeper.SetParams(s.Ctx(), bridgeParams))

	// Record supply from the staking module perspective
	supplyBefore, err := s.App.StakingKeeper.StakingTokenSupply(s.Ctx())
	s.Require().NoError(err)

	// Check that the supply reported by the staking module is the actual supply, which is different from
	// BridgeDenomTotalSupply. This legitimises our calculations below, which are based on BridgeDenomTotalSupply.
	s.Require().NoError(err)
	s.Require().False(supplyBefore.Equal(bridgeParams.BridgeDenomTotalSupply))

	// Run BeginBlocker a number of times and check that supply increases by a constant amount
	// and that it's transferring the minted tokens to the fee collector.
	for i := int64(1); i <= 10; i++ {

		beginBlockerCtx := s.Ctx()
		err = mint.BeginBlocker(beginBlockerCtx, s.App.MintKeeper, s.App.BridgeKeeper, inflationCalculationFn)
		s.Require().NoError(err)

		// Inflation = 0.1
		// BridgeDenomTotalSupply = 10000000000
		// AnnualProvisions = 10000000000 * 0.1
		//                  = 1000000000
		// Tokens minted per block = 1000000000 / 6311520
		//                         = 158.4404390701448
		//                         = 158

		// Check events
		s.AssertEventEmitted(beginBlockerCtx, minttypes.EventTypeMint, 1)
		event := s.FindEvent(beginBlockerCtx.EventManager().Events(), minttypes.EventTypeMint)
		eventAttributes := s.ExtractAttributes(event)
		s.Require().NotEmpty(eventAttributes[minttypes.AttributeKeyBondedRatio]) // the value is not important
		s.Require().Equal(eventAttributes[minttypes.AttributeKeyInflation], "0.100000000000000000")
		s.Require().Equal(eventAttributes[minttypes.AttributeKeyAnnualProvisions], "1000000000.000000000000000000")
		s.Require().Equal(eventAttributes[sdk.AttributeKeyAmount], "158")

		// Check minter attributes
		minter, err := s.App.MintKeeper.Minter.Get(s.Ctx())
		s.Require().NoError(err)
		s.Require().True(minter.Inflation.Equal(inflation))
		s.Require().True(minter.AnnualProvisions.Equal(sdkmath.LegacyMustNewDecFromStr("1000000000"))) // 0.1 * supply

		// Check fee collector balance is increasing by 158 each time
		feeCollectorBalance := s.App.BankKeeper.GetBalance(s.Ctx(), feeCollector, bondDenom)
		s.Require().True(feeCollectorBalance.Amount.Equal(sdkmath.NewInt(158 * i)))

		// Check supply is increasing by 158 each time
		supplyAfter, err := s.App.StakingKeeper.StakingTokenSupply(s.Ctx())
		s.Require().NoError(err)
		s.Require().True(supplyAfter.Equal(supplyBefore.AddRaw(158 * i)))
	}
}

// TestAppConfiguration_AppBeginBlockerRunsCustomMintLogic uses the App's BeginBlocker and tests that the correct
// mint module BeginBlocker is called. It does this just by ensuring there is only 1 mint event, and that the
// custom minting logic is being used, i.e. the inflation logic depends on the bridge module BridgeDenomTotalSupply.
func (s *MintModuleTestSuite) TestAppConfiguration_AppBeginBlockerRunsCustomMintLogic() {

	bondDenom := sdk.DefaultBondDenom

	// Simplify mint module params so that we have a constant 10% inflation.
	mintParams, err := s.App.MintKeeper.Params.Get(s.Ctx())
	s.Require().NoError(err)
	mintParams.InflationMin = sdkmath.LegacyMustNewDecFromStr("0.1")
	mintParams.InflationMax = sdkmath.LegacyMustNewDecFromStr("0.1")
	mintParams.InflationRateChange = sdkmath.LegacyMustNewDecFromStr("0.0")
	s.Require().NoError(s.App.MintKeeper.Params.Set(s.Ctx(), mintParams))

	// Override BridgeDenomTotalSupply so that we know what supply value will be used.
	bridgeParams := s.App.BridgeKeeper.GetParams(s.Ctx())
	bridgeParams.BridgeDenom = bondDenom
	bridgeParams.BridgeDenomTotalSupply = sdkmath.NewInt(10_000_000_000)
	s.Require().NoError(s.App.BridgeKeeper.SetParams(s.Ctx(), bridgeParams))

	// Record supply from the staking module perspective
	supplyBefore, err := s.App.StakingKeeper.StakingTokenSupply(s.Ctx())
	s.Require().NoError(err)

	bb, err := s.App.BeginBlocker(s.Ctx())
	s.Require().NoError(err)
	s.AssertEventInEventsList(bb.Events, minttypes.EventTypeMint, 1)

	// Inflation = 0.1
	// BridgeDenomTotalSupply = 10000000000
	// AnnualProvisions = 10000000000 * 0.1
	//                  = 1000000000
	// Tokens minted per block = 1000000000 / 6311520
	//                         = 158.4404390701448
	//                         = 158

	supplyAfter, err := s.App.StakingKeeper.StakingTokenSupply(s.Ctx())
	s.Require().NoError(err)
	s.Require().True(supplyAfter.Equal(supplyBefore.AddRaw(158)))
}
