package basic_test

import (
	"time"

	sdkmath "cosmossdk.io/math"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
)

func (s *BasicTestSuite) TestSlashedAmountSentToGovAccount() {
	s.Run("Set short signing window and slash fraction to 50%", func() {

		// 50% of every 10-block window has to be signed. Otherwise, the validator not signing will get slashed.
		slashingParams := s.QuerySlashingParams(s.Ctx())
		slashingParams.SignedBlocksWindow = int64(10)
		slashingParams.MinSignedPerWindow = sdkmath.LegacyMustNewDecFromStr("0.5")
		slashingParams.SlashFractionDowntime = sdkmath.LegacyMustNewDecFromStr("0.5")
		msgUpdateParams := slashingtypes.MsgUpdateParams{
			Authority: s.GetGovernanceAddress(),
			Params:    *slashingParams,
		}
		s.ExecuteGovProposal(&msgUpdateParams)
	})

	s.Run("Check that slashed funds due to downtime are sent to the governance account ", func() {

		initStake := testsuite.InitStakedCoin
		halfStake := sdk.NewCoin(initStake.Denom, initStake.Amount.QuoRaw(2))
		quarterStake := sdk.NewCoin(initStake.Denom, initStake.Amount.QuoRaw(4))

		// Sanity check Governance account balance pre-slash
		govAccount := s.GetGovernanceAddress()
		govAccountBalance, err := s.QueryBalance(s.Ctx(), govAccount, testsuite.BridgeDenom)
		s.Require().NoError(err)
		s.Require().True(govAccountBalance.Balance.IsZero())

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

		// Sanity check that tokens don't actually get burned
		govAccountBalance, err = s.QueryBalance(s.Ctx(), govAccount, testsuite.BridgeDenom)
		s.Require().NoError(err)
		s.Require().True(govAccountBalance.Balance.Equal(quarterStake))
	})
}
