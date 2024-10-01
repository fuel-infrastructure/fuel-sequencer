package basic_test

import (
	"strconv"
	"time"

	sdkmath "cosmossdk.io/math"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	cmtbytes "github.com/cometbft/cometbft/libs/bytes"
	"github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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

				// 10e27 Initial balance per Validator
				// 3 Validators
				// Total Supply = 30e27
				// BlocksPerYear = 6311520
				//
				// Inflation = 0.10
				//
				// Tokens minted per block = (1e27 / 6311520) * 0.10 where 1e27 is the BridgeDenomTotalSupply
				//                         = 158440439070144751185.134484244682738865
				//                         = 158440439070144751185
				//
				// First report will include total supply = 30e27 + (158440439070144751185 * 10) where 10 is the SupplyDeltaPeriod
				//                                        = 30000001584404390701447511850
				// Second report on will not include it   = (158440439070144751185 * 10) where 10 is the SupplyDeltaPeriod
				//                                        = 1584404390701447511850
				supply := testsuite.BridgeDenomTotalSupply
				params := minttypes.Params{BlocksPerYear: 6311520, MintDenom: testsuite.BridgeDenom}
				minter := minttypes.Minter{Inflation: sdkmath.LegacyMustNewDecFromStr("0.1")}
				minter.AnnualProvisions = minter.NextAnnualProvisions(params, supply)
				blockProvision := minter.BlockProvision(params).Amount
				supplyDeltaPeriodProvision := blockProvision.MulRaw(10)

				if expectedNonce == 1 {
					expectInitialSupply, ok := sdkmath.NewIntFromString("30000000000000000000000000000")
					s.Require().True(ok)
					expectReport := supplyDeltaPeriodProvision.Add(expectInitialSupply)
					s.Require().EqualValues(expectReport.String(), supplyDeltaString)
					s.Require().EqualValues(expectReport.String(), "30000001584404390701447511850")
				} else {
					expectReport := supplyDeltaPeriodProvision
					s.Require().EqualValues(expectReport.String(), supplyDeltaString)
					s.Require().EqualValues(expectReport.String(), "1584404390701447511850")
				}
				expectedNonce += 1
			}
		}
	})
}

func (s *BasicTestSuite) TestDowntimeSlashingAffectsSupplyDelta() {
	s.Run("Check that slashing due to downtime results in an offset in the supply delta", func() {

		supplyDeltaPeriod := s.QueryBridgeParams(s.Ctx()).SupplyDeltaPeriod
		supplyDeltaEvent, err := sdk.TypedEventToEvent(&bridgetypes.EventSupplyDeltaReported{})
		s.Require().NoError(err)

		initStake := testsuite.InitStakedCoin
		halfStake := sdk.NewCoin(initStake.Denom, initStake.Amount.QuoRaw(2))
		quarterStake := sdk.NewCoin(initStake.Denom, initStake.Amount.QuoRaw(4))

		// Reduce validator 0's stake so that when we shut it off, the chain proceeds without it
		_, err = s.SubmitMsgs(&stakingtypes.MsgUndelegate{
			DelegatorAddress: s.SeqKeys[0].AddressSeq,
			ValidatorAddress: s.SeqKeys[0].ValAddressSeq,
			Amount:           halfStake,
		})
		s.Require().NoError(err)
		s.PollForDelegationBalance(s.Ctx(), 10, s.SeqKeys[0].AddressSeq, s.SeqKeys[0].ValAddressSeq, halfStake)
		s.Require().NoError(s.WaitForSequencerBlocks(s.Ctx(), 1, time.Second*5))

		// Pause validator
		from, err := s.GetFuelSequencerHeight(s.Ctx())
		s.Require().NoError(err)
		s.PauseSequencer(0)

		// Wait enough time for enough blocks, for the validator to get slashed.
		// Note: we cannot wait for blocks because validator 0 is down.
		s.Sleep(time.Second * 15)

		// Unpause validator
		s.UnpauseSequencer(0)
		s.Sleep(time.Second * 5) // give some time for the validator to sync up
		until, err := s.GetFuelSequencerHeight(s.Ctx())
		s.Require().NoError(err)

		// Look for the slash event
		var slashEvent *abcitypes.Event
		var foundAt uint64
		for block := from; block <= until; block++ {
			event, found := s.SearchForEventInBlockResults(s.Ctx(), slashingtypes.EventTypeSlash, int64(block))
			if found {
				slashEvent = event
				foundAt = block
				break
			}
		}
		s.Require().NotZero(foundAt)

		// Check that half of the remaining stake (i.e. a quarter of the original) was burned
		slashAmount, ok := sdkmath.NewIntFromString(slashEvent.Attributes[4].Value)
		s.Require().True(ok)
		s.Require().Equal(slashAmount.String(), quarterStake.Amount.String())

		// Get the supply delta event where the slash was reported
		var supplyDeltaHeight int
		if foundAt%supplyDeltaPeriod == 0 {
			// Supply delta report is exactly at the slash event's height.
			supplyDeltaHeight = int(foundAt)
		} else {
			// Next supply delta report is some blocks in the future.
			supplyDeltaHeight = int(foundAt + supplyDeltaPeriod - (foundAt % supplyDeltaPeriod))
		}
		s.Require().NoError(s.WaitUntilSequencerBlock(s.Ctx(), supplyDeltaHeight, time.Minute))
		deltaEvent, found := s.SearchForEventInBlockResults(s.Ctx(), supplyDeltaEvent.Type, int64(supplyDeltaHeight))
		s.Require().True(found)
		supplyDelta := deltaEvent.Attributes[1].Value
		supplyDeltaAmount, ok := sdkmath.NewIntFromString(supplyDelta[1 : len(supplyDelta)-1])
		s.Require().True(ok)

		// 10e27 Initial balance per Validator
		// 3 Validators
		// Total Supply = 30e27
		// BlocksPerYear = 6311520
		//
		// Inflation = 0.10
		//
		// Tokens minted per block = (1e27 / 6311520) * 0.10 where 1e27 is the BridgeDenomTotalSupply
		//                         = 158440439070144751185.134484244682738865
		//                         = 158440439070144751185
		//
		// First report will include total supply = 30e27 + (158440439070144751185 * 10) where 10 is the SupplyDeltaPeriod
		//                                        = 30000001584404390701447511850
		// Second report on will not include it   = (158440439070144751185 * 10) where 10 is the SupplyDeltaPeriod
		//                                        = 1584404390701447511850
		//
		// From these reports we subtract the slash amount to get the expected supply delta amount.
		supply := testsuite.BridgeDenomTotalSupply
		params := minttypes.Params{BlocksPerYear: 6311520, MintDenom: testsuite.BridgeDenom}
		minter := minttypes.Minter{Inflation: sdkmath.LegacyMustNewDecFromStr("0.1")}
		minter.AnnualProvisions = minter.NextAnnualProvisions(params, supply)
		blockProvision := minter.BlockProvision(params).Amount
		supplyDeltaPeriodProvision := blockProvision.MulRaw(10)

		if supplyDeltaHeight == int(supplyDeltaPeriod) {
			expectInitialSupply, ok := sdkmath.NewIntFromString("30000000000000000000000000000")
			s.Require().True(ok)
			expectReport := supplyDeltaPeriodProvision.Add(expectInitialSupply).Sub(slashAmount)
			s.Require().EqualValues(expectReport.String(), supplyDeltaAmount.String())
		} else {
			expectReport := supplyDeltaPeriodProvision.Sub(slashAmount)
			s.Require().EqualValues(expectReport.String(), supplyDeltaAmount.String())
		}
	})
}
