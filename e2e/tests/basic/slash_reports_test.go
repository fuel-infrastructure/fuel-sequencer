package basic_test

import (
	"time"

	sdkmath "cosmossdk.io/math"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	reportstypes "github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"
)

func (s *BasicTestSuite) TestDowntimeSlashingRegistersSlashReport_PartialSlashMultipleDelegators() {
	s.Run("Check that slashing due to downtime results in a slash report getting generated", func() {

		initStake := testsuite.InitStakedCoin
		smallStake := sdk.NewInt64Coin(initStake.Denom, 10)
		halfStake := sdk.NewCoin(initStake.Denom, initStake.Amount.QuoRaw(2))
		quarterStake := sdk.NewCoin(initStake.Denom, initStake.Amount.QuoRaw(4))

		// Reduce validator 0's stake so that when we shut it off, the chain proceeds without it
		_, err := s.SubmitMsgsFromValidatorN(0, &stakingtypes.MsgUndelegate{
			DelegatorAddress: s.SeqKeys[0].AddressSeq,
			ValidatorAddress: s.SeqKeys[0].ValAddressSeq,
			Amount:           halfStake,
		})
		s.Require().NoError(err)
		s.PollForDelegationBalance(s.Ctx(), 10, s.SeqKeys[0].AddressSeq, s.SeqKeys[0].ValAddressSeq, halfStake)

		// Delegate some tokens from another address so that we get more interesting results
		_, err = s.SubmitMsgsFromValidatorN(1, &stakingtypes.MsgDelegate{
			DelegatorAddress: s.SeqKeys[1].AddressSeq,
			ValidatorAddress: s.SeqKeys[0].ValAddressSeq, // delegate to validator 0 from user 1
			Amount:           smallStake,
		})
		s.Require().NoError(err)
		s.PollForDelegationBalance(s.Ctx(), 10, s.SeqKeys[1].AddressSeq, s.SeqKeys[0].ValAddressSeq, smallStake)

		// Wait for delegations to take effect
		s.Require().NoError(s.WaitForSequencerBlocks(s.Ctx(), 2, time.Second*10))

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

		// Check that there is a slash report at the slash height
		slashReport := s.QuerySlashReport(s.Ctx(), foundAt)

		// Original values:
		// - Validator shares (VS) = 1000000010
		// - Validator tokens (VT) = 1000000010
		// - Delegator0 shares (D0S) = 1000000000
		// - Delegator1 shares (D1S) = 10
		//
		// During slash (Cosmos SDK side):
		// - Slash factor (SF) = 0.5
		// - Validator consensus power (VCP) = 1
		// - Validator tokens from consensus power (VTCP): VCP x 1e9 = 1000000000
		// - Slash amount (SA) = Truncate(VTCP * SF) = 500000000
		// - Effective fraction (EF) = QuoRoundUp(SA, VT) = 0.499999995000000050
		//
		// During slash (Reports module side):
		// - Delegator0 tokens from shares (D0TFS) = (D0S * VT) / VS = 1000000000
		// - Delegator1 tokens from shares (D1TFS) = (D1S * VT) / VS = 10
		// - Delegator0 slash = Truncate(D0TFS * EF) = 499999995
		// - Delegator1 slash = Truncate(D1TFS * EF) = 4
		//
		// Post slash:
		// - New validator tokens (NVT) = 500000010
		// - Delegator0 tokens from shares = Truncate((D0S * NVT) / VS) = 500000004
		// - Delegator1 tokens from shares = Truncate((D1S * NVT) / VS) = 5
		expectedSlashReport := &reportstypes.SlashReport{
			Height: foundAt,
			Entries: []reportstypes.SlashEntry{
				{
					ValidatorAddress:          s.SeqKeys[0].ValAddressSeq,
					DelegatorAddress:          s.SeqKeys[0].AddressSeq, // user 0
					DelegatorSlashAmount:      sdkmath.NewInt(499999995),
					DelegatorBondedBalance:    sdkmath.NewInt(500000004),
					DelegatorUnbondingBalance: halfStake.Amount,
				},
				{
					ValidatorAddress:          s.SeqKeys[0].ValAddressSeq,
					DelegatorAddress:          s.SeqKeys[1].AddressSeq, // user 1
					DelegatorSlashAmount:      sdkmath.NewInt(4),
					DelegatorBondedBalance:    sdkmath.NewInt(5),
					DelegatorUnbondingBalance: sdkmath.ZeroInt(),
				},
			},
		}
		s.Require().EqualValues(expectedSlashReport, slashReport)
	})
}

func (s *BasicTestSuite) TestDowntimeSlashingRegistersSlashReport_FullSlashMultipleDelegators() {
	s.Run("Change slash fraction to 100%", func() {

		// Create a new dummy proposal to vote on
		slashingParams := s.QuerySlashingParams(s.Ctx())
		slashingParams.SlashFractionDowntime = sdkmath.LegacyOneDec()
		msgUpdateParams := slashingtypes.MsgUpdateParams{
			Authority: s.GetGovernanceAddress(),
			Params:    *slashingParams,
		}
		s.ExecuteGovProposal(&msgUpdateParams)
	})

	s.Run("Check that slashing due to downtime results in a slash report getting generated", func() {

		initStake := testsuite.InitStakedCoin
		smallStake := sdk.NewInt64Coin(initStake.Denom, 10)
		halfStake := sdk.NewCoin(initStake.Denom, initStake.Amount.QuoRaw(2))

		// Reduce validator 0's stake so that when we shut it off, the chain proceeds without it
		_, err := s.SubmitMsgsFromValidatorN(0, &stakingtypes.MsgUndelegate{
			DelegatorAddress: s.SeqKeys[0].AddressSeq,
			ValidatorAddress: s.SeqKeys[0].ValAddressSeq,
			Amount:           halfStake,
		})
		s.Require().NoError(err)
		s.PollForDelegationBalance(s.Ctx(), 10, s.SeqKeys[0].AddressSeq, s.SeqKeys[0].ValAddressSeq, halfStake)

		// Delegate some tokens from another address so that we get more interesting results
		_, err = s.SubmitMsgsFromValidatorN(1, &stakingtypes.MsgDelegate{
			DelegatorAddress: s.SeqKeys[1].AddressSeq,
			ValidatorAddress: s.SeqKeys[0].ValAddressSeq, // delegate to validator 0 from user 1
			Amount:           smallStake,
		})
		s.Require().NoError(err)
		s.PollForDelegationBalance(s.Ctx(), 10, s.SeqKeys[1].AddressSeq, s.SeqKeys[0].ValAddressSeq, smallStake)

		// Wait for delegations to take effect
		s.Require().NoError(s.WaitForSequencerBlocks(s.Ctx(), 2, time.Second*10))

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

		// Check that all the remaining stake (i.e. a half of the original) was burned
		slashAmount, ok := sdkmath.NewIntFromString(slashEvent.Attributes[4].Value)
		s.Require().True(ok)
		s.Require().Equal(slashAmount.String(), halfStake.Amount.String())

		// Check that there is a slash report at the slash height
		slashReport := s.QuerySlashReport(s.Ctx(), foundAt)

		// Original values:
		// - Validator shares (VS) = 1000000010
		// - Validator tokens (VT) = 1000000010
		// - Delegator0 shares (D0S) = 1000000000
		// - Delegator1 shares (D1S) = 10
		//
		// During slash (Cosmos SDK side):
		// - Slash factor (SF) = 1.0
		// - Validator consensus power (VCP) = 1
		// - Validator tokens from consensus power (VTCP): VCP x 1e9 = 1000000000
		// - Slash amount (SA) = Truncate(VTCP * SF) = 1000000000
		// - Effective fraction (EF) = QuoRoundUp(SA, VT) = 0.999999990000000100
		//
		// During slash (Reports module side):
		// - Delegator0 tokens from shares (D0TFS) = (D0S * VT) / VS = 1000000000
		// - Delegator1 tokens from shares (D1TFS) = (D1S * VT) / VS = 10
		// - Delegator0 slash = Truncate(D0TFS * EF) = 999999990
		// - Delegator1 slash = Truncate(D1TFS * EF) = 9
		//
		// Post slash:
		// - New validator tokens (NVT) = 10
		// - Delegator0 tokens from shares = Truncate((D0S * NVT) / VS) = 9
		// - Delegator1 tokens from shares = Truncate((D1S * NVT) / VS) = 0
		expectedSlashReport := &reportstypes.SlashReport{
			Height: foundAt,
			Entries: []reportstypes.SlashEntry{
				{
					ValidatorAddress:          s.SeqKeys[0].ValAddressSeq,
					DelegatorAddress:          s.SeqKeys[0].AddressSeq, // user 0
					DelegatorSlashAmount:      sdkmath.NewInt(999999990),
					DelegatorBondedBalance:    sdkmath.NewInt(9),
					DelegatorUnbondingBalance: halfStake.Amount,
				},
				{
					ValidatorAddress:          s.SeqKeys[0].ValAddressSeq,
					DelegatorAddress:          s.SeqKeys[1].AddressSeq, // user 1
					DelegatorSlashAmount:      sdkmath.NewInt(9),
					DelegatorBondedBalance:    sdkmath.ZeroInt(),
					DelegatorUnbondingBalance: sdkmath.ZeroInt(),
				},
			},
		}
		s.Require().EqualValues(expectedSlashReport, slashReport)
	})
}

func (s *BasicTestSuite) TestDowntimeSlashingRegistersSlashReport_FullSlashOneDelegator() {
	s.Run("Change slash fraction to 100%", func() {

		// Create a new dummy proposal to vote on
		slashingParams := s.QuerySlashingParams(s.Ctx())
		slashingParams.SlashFractionDowntime = sdkmath.LegacyOneDec()
		msgUpdateParams := slashingtypes.MsgUpdateParams{
			Authority: s.GetGovernanceAddress(),
			Params:    *slashingParams,
		}
		s.ExecuteGovProposal(&msgUpdateParams)
	})

	s.Run("Check that slashing due to downtime results in a slash report getting generated", func() {

		initStake := testsuite.InitStakedCoin
		halfStake := sdk.NewCoin(initStake.Denom, initStake.Amount.QuoRaw(2))
		twiceStake := sdk.NewCoin(initStake.Denom, initStake.Amount.MulRaw(2))

		// Increase validator 1's and 2's stake so that when we shut off validator 0, the chain proceeds without it
		_, err := s.SubmitMsgsFromValidatorN(1, &stakingtypes.MsgDelegate{
			DelegatorAddress: s.SeqKeys[1].AddressSeq,
			ValidatorAddress: s.SeqKeys[1].ValAddressSeq,
			Amount:           initStake,
		})
		s.Require().NoError(err)
		s.PollForDelegationBalance(s.Ctx(), 10, s.SeqKeys[1].AddressSeq, s.SeqKeys[1].ValAddressSeq, twiceStake)
		_, err = s.SubmitMsgsFromValidatorN(2, &stakingtypes.MsgDelegate{
			DelegatorAddress: s.SeqKeys[2].AddressSeq,
			ValidatorAddress: s.SeqKeys[2].ValAddressSeq,
			Amount:           initStake,
		})
		s.Require().NoError(err)
		s.PollForDelegationBalance(s.Ctx(), 10, s.SeqKeys[2].AddressSeq, s.SeqKeys[2].ValAddressSeq, twiceStake)

		// Wait for delegations to take effect
		s.Require().NoError(s.WaitForSequencerBlocks(s.Ctx(), 2, time.Second*10))

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

		// Check that all the initial stake was burned
		slashAmount, ok := sdkmath.NewIntFromString(slashEvent.Attributes[4].Value)
		s.Require().True(ok)
		s.Require().Equal(slashAmount.String(), initStake.Amount.String())

		// Check that there is a slash report at the slash height
		slashReport := s.QuerySlashReport(s.Ctx(), foundAt)

		// Original values:
		// - Validator shares (VS) = 2000000000
		// - Validator tokens (VT) = 2000000000
		// - Delegator0 shares (D0S) = 2000000000
		//
		// During slash (Cosmos SDK side):
		// - Slash factor (SF) = 1.0
		// - Validator consensus power (VCP) = 1
		// - Validator tokens from consensus power (VTCP): VCP x 1e9 = 2000000000
		// - Slash amount (SA) = Truncate(VTCP * SF) = 2000000000
		// - Effective fraction (EF) = QuoRoundUp(SA, VT) = 1.0
		//
		// During slash (Reports module side):
		// - Delegator0 tokens from shares (D0TFS) = (D0S * VT) / VS = 2000000000
		// - Delegator0 slash = Truncate(D0TFS * EF) = 2000000000
		//
		// Post slash:
		// - New validator tokens (NVT) = 0
		// - Delegator0 tokens from shares = Truncate((D0S * NVT) / VS) = 0
		expectedSlashReport := &reportstypes.SlashReport{
			Height: foundAt,
			Entries: []reportstypes.SlashEntry{
				{
					ValidatorAddress:          s.SeqKeys[0].ValAddressSeq,
					DelegatorAddress:          s.SeqKeys[0].AddressSeq,
					DelegatorSlashAmount:      initStake.Amount,
					DelegatorBondedBalance:    sdkmath.ZeroInt(),
					DelegatorUnbondingBalance: halfStake.Amount,
				},
			},
		}
		s.Require().EqualValues(expectedSlashReport, slashReport)
	})
}
