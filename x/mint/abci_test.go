package mint_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/mint"
)

func (s *MintModuleTestSuite) TestBeginBlocker_InflationBasedOnBridgeModuleParams() {

	bondDenom := sdk.DefaultBondDenom
	inflation := sdkmath.LegacyMustNewDecFromStr("0.1")
	feeCollector := s.App.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)

	// Set the inflation rate via the minter even though this will get overridden by InflationMin and InflationMax
	minter, err := s.App.MintKeeper.Minter.Get(s.Ctx())
	s.Require().NoError(err)
	minter.Inflation = inflation
	s.Require().NoError(s.App.MintKeeper.Minter.Set(s.Ctx(), minter))

	// Simplify mint module params so that inflation is set to zero upon the first call to the custom mint BeginBlocker.
	mintParams, err := s.App.MintKeeper.Params.Get(s.Ctx())
	s.Require().NoError(err)
	mintParams.InflationMin = sdkmath.LegacyMustNewDecFromStr("0.0")        // sets inflation to 0
	mintParams.InflationMax = sdkmath.LegacyMustNewDecFromStr("0.0")        // sets inflation to 0
	mintParams.InflationRateChange = sdkmath.LegacyMustNewDecFromStr("0.0") // this is not used
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

	// Run BeginBlocker once to ensure that the minter inflation changes based on InflationMin and InflationMax.
	beginBlockerCtx := s.Ctx()
	err = mint.BeginBlocker(beginBlockerCtx, s.App.MintKeeper, s.App.BridgeKeeper, s.App.BondKeeper)
	s.Require().NoError(err)
	// ...check inflation and annual provisions are zero
	minter, err = s.App.MintKeeper.Minter.Get(s.Ctx())
	s.Require().NoError(err)
	s.Require().True(minter.Inflation.IsZero())
	s.Require().True(minter.AnnualProvisions.IsZero())

	// Run BeginBlocker a number of times and check that supply does not increase when inflation is zero.
	for i := int64(1); i <= 10; i++ {

		beginBlockerCtx := s.Ctx()
		err = mint.BeginBlocker(beginBlockerCtx, s.App.MintKeeper, s.App.BridgeKeeper, s.App.BondKeeper)
		s.Require().NoError(err)

		// Check events
		s.AssertEventEmitted(beginBlockerCtx, minttypes.EventTypeMint, 1)
		event := s.FindEvent(beginBlockerCtx.EventManager().Events(), minttypes.EventTypeMint)
		eventAttributes := s.ExtractAttributes(event)
		s.Require().NotEmpty(eventAttributes[minttypes.AttributeKeyBondedRatio]) // the value is not important
		s.Require().Equal(eventAttributes[minttypes.AttributeKeyInflation], "0.000000000000000000")
		s.Require().Equal(eventAttributes[minttypes.AttributeKeyAnnualProvisions], "0.000000000000000000")
		s.Require().Equal(eventAttributes[sdk.AttributeKeyAmount], "0")

		// Check minter attributes are still zero
		minter, err := s.App.MintKeeper.Minter.Get(s.Ctx())
		s.Require().NoError(err)
		s.Require().True(minter.Inflation.IsZero())
		s.Require().True(minter.AnnualProvisions.IsZero())
	}

	// Set non-zero inflation of 0.1 by changing InflationMin and InflationMax
	mintParams.InflationMin = inflation
	mintParams.InflationMax = inflation
	s.Require().NoError(s.App.MintKeeper.Params.Set(s.Ctx(), mintParams))

	// Run BeginBlocker a number of times and check that supply increases by a constant amount
	// and that it's transferring the minted tokens to the fee collector.
	for i := int64(1); i <= 10; i++ {

		beginBlockerCtx := s.Ctx()
		err = mint.BeginBlocker(beginBlockerCtx, s.App.MintKeeper, s.App.BridgeKeeper, s.App.BondKeeper)
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

		// Check supply is increasing by 158 each time, confirming that the custom minting logic is being used
		supplyAfter, err := s.App.StakingKeeper.StakingTokenSupply(s.Ctx())
		s.Require().NoError(err)
		s.Require().True(supplyAfter.Equal(supplyBefore.AddRaw(158 * i)))
	}
}

func (s *MintModuleTestSuite) TestBeginBlocker_InflationBasedOnInflationMinMax() {

	zeroInflation := sdkmath.LegacyZeroDec()
	nonZeroInflation := sdkmath.LegacyMustNewDecFromStr("0.1")
	initialInflation := sdkmath.LegacyMustNewDecFromStr("0.5") // arbitrary value to be able to detect changes

	testCases := []struct {
		name              string
		inflationMin      sdkmath.LegacyDec
		inflationMax      sdkmath.LegacyDec
		expectedInflation sdkmath.LegacyDec
	}{
		{
			name:              "equal inflation params (0) => inflation updated",
			inflationMin:      zeroInflation,
			inflationMax:      zeroInflation,
			expectedInflation: zeroInflation, // updated
		},
		{
			name:              "equal inflation params (0.1) => inflation updated",
			inflationMin:      nonZeroInflation,
			inflationMax:      nonZeroInflation,
			expectedInflation: nonZeroInflation, // updated
		},
		{
			name:              "unequal inflation params => inflation unchanged",
			inflationMin:      zeroInflation,
			inflationMax:      nonZeroInflation,
			expectedInflation: initialInflation, // unchanged
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Setup()

			// Set initial inflation rate
			minter, err := s.App.MintKeeper.Minter.Get(s.Ctx())
			s.Require().NoError(err)
			minter.Inflation = initialInflation
			s.Require().NoError(s.App.MintKeeper.Minter.Set(s.Ctx(), minter))

			// Set inflation min and max params
			params, err := s.App.MintKeeper.Params.Get(s.Ctx())
			s.Require().NoError(err)
			params.InflationMin = tc.inflationMin
			params.InflationMax = tc.inflationMax
			s.Require().NoError(s.App.MintKeeper.Params.Set(s.Ctx(), params))

			// Run BeginBlocker
			err = mint.BeginBlocker(s.Ctx(), s.App.MintKeeper, s.App.BridgeKeeper, s.App.BondKeeper)
			s.Require().NoError(err)

			// Check inflation rate
			minter, err = s.App.MintKeeper.Minter.Get(s.Ctx())
			s.Require().NoError(err)
			s.Require().True(minter.Inflation.Equal(tc.expectedInflation))
		})
	}
}

// TestAppConfiguration_AppBeginBlockerRunsCustomMintLogic uses the App's BeginBlocker instead of the mint module one
// directly. It tests that the correct mint module BeginBlocker is called, ensuring that we've correctly wired-up the
// app. It does this just by ensuring there is only 1 mint event, and that the custom minting logic is being used.
func (s *MintModuleTestSuite) TestAppConfiguration_AppBeginBlockerRunsCustomMintLogic() {

	bondDenom := sdk.DefaultBondDenom

	// Set the inflation rate via the minter
	minter, err := s.App.MintKeeper.Minter.Get(s.Ctx())
	s.Require().NoError(err)
	minter.Inflation = sdkmath.LegacyMustNewDecFromStr("0.1")
	s.Require().NoError(s.App.MintKeeper.Minter.Set(s.Ctx(), minter))

	// Simplify mint module params so that we have a constant 10% inflation.
	mintParams, err := s.App.MintKeeper.Params.Get(s.Ctx())
	s.Require().NoError(err)
	mintParams.InflationMin = minter.Inflation                              // sets inflation to minter.Inflation
	mintParams.InflationMax = minter.Inflation                              // sets inflation to minter.Inflation
	mintParams.InflationRateChange = sdkmath.LegacyMustNewDecFromStr("0.0") // this is not used
	s.Require().NoError(s.App.MintKeeper.Params.Set(s.Ctx(), mintParams))

	// Override BridgeDenomTotalSupply so that we know what supply value will be used.
	bridgeParams := s.App.BridgeKeeper.GetParams(s.Ctx())
	bridgeParams.BridgeDenom = bondDenom
	bridgeParams.BridgeDenomTotalSupply = sdkmath.NewInt(10_000_000_000)
	s.Require().NoError(s.App.BridgeKeeper.SetParams(s.Ctx(), bridgeParams))

	// Record supply from the staking module perspective
	supplyBefore, err := s.App.StakingKeeper.StakingTokenSupply(s.Ctx())
	s.Require().NoError(err)

	// Use app's BeginBlocker to ensure the app is correctly wired-up with the custom minting logic.
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

	// Check supply is increasing by 158 each time, confirming that the custom minting logic is being used
	supplyAfter, err := s.App.StakingKeeper.StakingTokenSupply(s.Ctx())
	s.Require().NoError(err)
	s.Require().True(supplyAfter.Equal(supplyBefore.AddRaw(158)))
}

func (s *MintModuleTestSuite) TestBeginBlocker_CoinDistribution() {
	bondDenom := sdk.DefaultBondDenom
	feeCollector := s.App.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
	bondAuthority := s.App.AccountKeeper.GetModuleAddress(govtypes.ModuleName)

	totalSupply := types.TestBridgeDenomTotalSupply
	blocksPerYear := types.BlocksPerYear
	blocksPerYearInt := sdkmath.NewIntFromUint64(blocksPerYear)
	blocksPerYearLegacyInt := sdkmath.LegacyNewDecFromInt(blocksPerYearInt)

	calculateExpectedAmounts := func(mintInflation, bondInflation string) (
		mintedAmount, feeAmount, bondAmount int64,
	) {
		mintInflationDec, err := sdkmath.LegacyNewDecFromStr(mintInflation)
		s.Require().NoError(err)
		bondInflationDec, err := sdkmath.LegacyNewDecFromStr(bondInflation)
		s.Require().NoError(err)

		totalInflation := mintInflationDec.Add(bondInflationDec)
		if totalInflation.IsZero() {
			return 0, 0, 0
		}

		// Calculate total minted amount
		inflationPerYear := totalSupply.ToLegacyDec().Mul(totalInflation)
		mintedAmount = inflationPerYear.Quo(blocksPerYearLegacyInt).TruncateInt().Int64()

		// Calculate split between fee collector and bond authority
		mintRatio := mintInflationDec.Quo(totalInflation)
		feeAmount = sdkmath.LegacyNewDecFromInt(sdkmath.NewInt(mintedAmount)).Mul(mintRatio).TruncateInt().Int64()
		bondAmount = mintedAmount - feeAmount

		return mintedAmount, feeAmount, bondAmount
	}

	testCases := []struct {
		name          string
		mintInflation string // as decimal string
		bondInflation string // as decimal string
	}{
		{
			name:          "zero mint inflation, zero bond inflation",
			mintInflation: "0.0",
			bondInflation: "0.0",
		},
		{
			name:          "zero mint inflation, non-zero bond inflation",
			mintInflation: "0.0",
			bondInflation: "0.1",
		},
		{
			name:          "non-zero mint inflation, zero bond inflation",
			mintInflation: "0.1",
			bondInflation: "0.0",
		},
		{
			name:          "equal mint and bond inflation",
			mintInflation: "0.1",
			bondInflation: "0.1",
		},
		{
			name:          "mint inflation double bond inflation",
			mintInflation: "0.2",
			bondInflation: "0.1",
		},
		{
			name:          "bond inflation double mint inflation",
			mintInflation: "0.1",
			bondInflation: "0.2",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Setup()

			// Set up bridge params with fixed total supply for predictable calculations
			bridgeParams := s.App.BridgeKeeper.GetParams(s.Ctx())
			bridgeParams.BridgeDenom = bondDenom
			bridgeParams.BridgeDenomTotalSupply = totalSupply
			s.Require().NoError(s.App.BridgeKeeper.SetParams(s.Ctx(), bridgeParams))

			// Set mint inflation via InflationMin/Max params
			mintParams, err := s.App.MintKeeper.Params.Get(s.Ctx())
			s.Require().NoError(err)
			mintParams.InflationMin = sdkmath.LegacyMustNewDecFromStr(tc.mintInflation)
			mintParams.InflationMax = mintParams.InflationMin
			mintParams.BlocksPerYear = blocksPerYear
			mintParams.MintDenom = bondDenom // This matches the bridge denom
			s.Require().NoError(s.App.MintKeeper.Params.Set(s.Ctx(), mintParams))

			// Set bond inflation
			bondParams := s.App.BondKeeper.GetParams(s.Ctx())
			bondParams.Inflation = sdkmath.LegacyMustNewDecFromStr(tc.bondInflation)
			s.Require().NoError(s.App.BondKeeper.SetParams(s.Ctx(), bondParams))

			// Calculate expected values
			expectMintedAmount, expectFeeAmount, expectBondAmount := calculateExpectedAmounts(tc.mintInflation, tc.bondInflation)

			// Run BeginBlocker
			beginBlockerCtx := s.Ctx()
			err = mint.BeginBlocker(beginBlockerCtx, s.App.MintKeeper, s.App.BridgeKeeper, s.App.BondKeeper)
			s.Require().NoError(err)

			// Check minted amount in event
			s.AssertEventEmitted(beginBlockerCtx, minttypes.EventTypeMint, 1)
			event := s.FindEvent(beginBlockerCtx.EventManager().Events(), minttypes.EventTypeMint)
			eventAttributes := s.ExtractAttributes(event)
			s.Require().Equal(fmt.Sprintf("%d", expectMintedAmount), eventAttributes[sdk.AttributeKeyAmount])

			// Check fee collector balance
			feeCollectorBalance := s.App.BankKeeper.GetBalance(s.Ctx(), feeCollector, bondDenom)
			s.Require().True(feeCollectorBalance.Amount.Equal(sdkmath.NewInt(expectFeeAmount)),
				"fee collector balance: expected %d, got %d", expectFeeAmount, feeCollectorBalance.Amount.Int64())

			// Check bond authority balance
			bondAuthorityBalance := s.App.BankKeeper.GetBalance(s.Ctx(), bondAuthority, bondDenom)
			s.Require().True(bondAuthorityBalance.Amount.Equal(sdkmath.NewInt(expectBondAmount)),
				"bond authority balance: expected %d, got %d", expectBondAmount, bondAuthorityBalance.Amount.Int64())
		})
	}
}
