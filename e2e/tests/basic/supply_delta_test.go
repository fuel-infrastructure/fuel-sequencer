package basic_test

import (
	"strconv"
	"time"

	sdkmath "cosmossdk.io/math"
	cmtbytes "github.com/cometbft/cometbft/libs/bytes"
	"github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *BasicTestSuite) TestMsgSupplyDeltaIsInjected() {
	s.Run("Check that MsgSupplyDelta is getting injected only at the right heights", func() {

		supplyDeltaPeriod := int64(s.QueryBridgeParams(s.Ctx()).SupplyDeltaPeriod)
		supplyDeltaEvent, err := sdk.TypedEventToEvent(&bridgetypes.EventSupplyDeltaReported{})
		s.Require().NoError(err)

		// Wait for at least two MsgSupplyDelta to get injected
		searchTillBlock := supplyDeltaPeriod * 2
		err = s.WaitForSequencerBlocks(s.Ctx(), int(searchTillBlock), time.Minute)
		s.Require().NoError(err)

		// Construct standard MsgSupplyDelta details
		typicalMsgSupplyDelta := bridgetypes.MsgSupplyDelta{Authority: s.GetGovernanceAddress()}
		msgSupplyDeltaTypeUrl := sdk.MsgTypeURL(&bridgetypes.MsgSupplyDelta{})

		// To ensure that all MsgSupplyDelta transaction hashes are unique.
		uniqueHashes := make(map[string]bool)

		// Ensure that MsgSupplyDelta injected, and only at the right heights.
		expectedNonce := uint64(1)
		for block := int64(1); block <= searchTillBlock; block++ {
			expectMsgSupplyDelta := block%supplyDeltaPeriod == 0

			// Search for message showing successful injection
			msg, txBz, msgFound := s.SearchForMsgInBlock(s.Ctx(), msgSupplyDeltaTypeUrl, block)
			s.Require().Equal(expectMsgSupplyDelta, msgFound)
			if expectMsgSupplyDelta {
				msgSupplyDelta, ok := msg.(*bridgetypes.MsgSupplyDelta)
				s.Require().True(ok)
				s.Require().EqualValues(typicalMsgSupplyDelta, *msgSupplyDelta)

				// Track hashes to ensure that all of them are unique.
				txHash := cmtbytes.HexBytes(types.Tx(txBz).Hash()).String()
				if _, alreadySeen := uniqueHashes[txHash]; alreadySeen {
					s.Require().False(alreadySeen)
				}
				uniqueHashes[txHash] = true
			}

			// Search for event showing successful execution
			event, eventFound := s.SearchForEventInBlockResults(s.Ctx(), supplyDeltaEvent.Type, block)
			s.Require().Equal(expectMsgSupplyDelta, eventFound)

			if expectMsgSupplyDelta {
				nonceAttribute := event.Attributes[0]
				s.Require().EqualValues("nonce", nonceAttribute.Key)
				nonceValue := nonceAttribute.Value[1 : len(nonceAttribute.Value)-1]
				s.Require().EqualValues(strconv.FormatUint(expectedNonce, 10), nonceValue)

				supplyDeltaAttribute := event.Attributes[1]
				s.Require().EqualValues("supply_delta", supplyDeltaAttribute.Key)
				supplyDeltaString := supplyDeltaAttribute.Value[1 : len(supplyDeltaAttribute.Value)-1]

				// 20e18 Initial balance per Validator
				// 3 Validators
				// Total Supply = 60e18
				// BlocksPerYear = 31557600
				//
				// Inflation = 0.10
				//
				// Tokens minted per block = (1e19 / 31557600) * 0.10 where 1e19 is the BridgeDenomTotalSupply
				//                         = 31688087814.02895023703
				//                         = 31688087814
				//
				// First report will include total supply = 60e18 + (31688087814 * 10) where 10 is the SupplyDeltaPeriod
				//                                        = 60000000316880878140
				// Second report on will not include it   = (31688087814 * 10) where 10 is the SupplyDeltaPeriod
				//                                        = 316880878140
				supply := testsuite.BridgeDenomTotalSupply
				params := minttypes.Params{BlocksPerYear: 6311520, MintDenom: testsuite.BridgeDenom}
				minter := minttypes.Minter{Inflation: sdkmath.LegacyMustNewDecFromStr("0.1")}
				minter.AnnualProvisions = minter.NextAnnualProvisions(params, supply)
				blockProvision := minter.BlockProvision(params).Amount
				supplyDeltaPeriodProvision := blockProvision.MulRaw(10)

				if expectedNonce == 1 {
					expectInitialSupply, ok := sdkmath.NewIntFromString("60000000000000000000")
					s.Require().True(ok)
					expectReport := supplyDeltaPeriodProvision.Add(expectInitialSupply)
					s.Require().EqualValues(expectReport.String(), supplyDeltaString)
					s.Require().EqualValues(expectReport.String(), "60000000316880878140")
				} else {
					expectReport := supplyDeltaPeriodProvision
					s.Require().EqualValues(expectReport.String(), supplyDeltaString)
					s.Require().EqualValues(expectReport.String(), "316880878140")
				}
				expectedNonce += 1
			}
		}
	})
}
