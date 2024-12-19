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

func (s *BasicTestSuite) TestDowntimeSlashingRegistersSlashReport() {
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

		expectedSlashReport := reportstypes.SlashReport{
			Height: 0,
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
					DelegatorUnbondingBalance: sdkmath.Int{},
				},
			},
		}
		s.Require().Equal(expectedSlashReport, slashReport)

		// BEFORE
		// fuelsequencer1vtfzrk6f4m6kxt6ehyqt9j5su5hvcz5q3dmlsm : 10test
		// fuelsequencer15yk64u7zc9g9k2yr2wmzeva5qgwxps6y3z4xeu : 1000000000test
		//
		// TOTAL: 1000000010test
		//
		// AFTER
		// fuelsequencer1vtfzrk6f4m6kxt6ehyqt9j5su5hvcz5q3dmlsm : 5test			(slash: 4)
		// fuelsequencer15yk64u7zc9g9k2yr2wmzeva5qgwxps6y3z4xeu : 500000004test (slash: 499999995)
		//
		// TOTAL: 1000000008test

		// 2024-12-19 16:56:27 3:56PM WRN CustomBeforeValidatorSlashed :: INSERTING SLASH ENTRY amt=4 del=fuelsequencer1vtfzrk6f4m6kxt6ehyqt9j5su5hvcz5q3dmlsm fraction=0.499999995000000050 module=server tokens_from_shares=10.000000000000000000
		// 2024-12-19 16:56:27 3:56PM WRN CustomBeforeValidatorSlashed :: INSERTING SLASH ENTRY amt=499999995 del=fuelsequencer15yk64u7zc9g9k2yr2wmzeva5qgwxps6y3z4xeu fraction=0.499999995000000050 module=server tokens_from_shares=1000000000.000000000000000000
	})
}
