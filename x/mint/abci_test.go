package mint_test

import (
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

func (s *MintModuleTestSuite) TestBeginBlocker_ExpectsBridgeDenomEqualMintDenom() {
	_, err := s.BeginBlock()
	s.Require().ErrorContains(err, "mismatching bridge and mint denoms: ufuel != stake")
}

func (s *MintModuleTestSuite) TestBeginBlocker_InflationBasedOnBridgeModuleParams() {

	bondDenom := types.DefaultBondDenom

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

	// Check that the supply reported by the staking module is the actual supply, which is different from
	// BridgeDenomTotalSupply. This legitimises our calculations below, which are based on BridgeDenomTotalSupply.
	s.Require().NoError(err)
	s.Require().False(supplyBefore.Equal(bridgeParams.BridgeDenomTotalSupply))

	// Run BeginBlocker a number of times and check that supply increases by a constant amount
	// and that it's transferring the minted tokens to the fee collector.
	for i := int64(1); i <= 10; i++ {

		bb, err := s.BeginBlock()
		s.Require().NoError(err)
		s.AssertEventInEventsList(bb.Events, minttypes.EventTypeMint, 1)

		supplyAfter, err := s.App.StakingKeeper.StakingTokenSupply(s.Ctx())
		s.Require().NoError(err)

		// Inflation = 0.1
		// BridgeDenomTotalSupply = 10000000000
		// Tokens minted per block = (10000000000 / 6311520) * 0.1
		//                         = 158.440439070144751
		//                         = 158

		s.Require().True(supplyAfter.Equal(supplyBefore.AddRaw(158 * i)))
	}
}
