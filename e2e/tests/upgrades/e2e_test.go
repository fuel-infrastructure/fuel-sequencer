package upgrades_test

import (
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/fuel-infrastructure/fuel-sequencer/app/upgrades/power_reduction"
	e2etestsuite "github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

const (
	haltHeightDelta    = uint64(25) // will propose upgrade this many blocks in the future; must be > voting period
	blocksAfterUpgrade = uint64(10) // will wait for this many blocks after the upgrade
	upgradeName        = power_reduction.UpgradeName
	fromImageVersion   = "b651895"                               // this image needs to exist for this test to run
	toImageVersion     = "hotfix_adjust-default-power-reduction" // this image needs to exist for this test to run
)

type UpgradesTestSuite struct {
	e2etestsuite.E2ETestSuite
}

func TestUpgradesTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradesTestSuite))
}

func (s *UpgradesTestSuite) SetupTest() {

	s.FuelSequencerDockerImageTag = fromImageVersion

	s.E2ETestSuite.SetupTest()
}

func (s *UpgradesTestSuite) TestUpgradePowerReduction() {

	s.Run("Perform the upgrade", func() {
		height, err := s.GetFuelSequencerHeight(s.Ctx())
		s.Require().NoError(err, "error fetching height before submit upgrade proposal")

		haltHeight := height + haltHeightDelta
		s.Logger().Info("Submitting software upgrade proposal", zap.Uint64("halt_height", haltHeight))
		msgUpgrade := &upgradetypes.MsgSoftwareUpgrade{
			Authority: s.GetGovernanceAddress(),
			Plan: upgradetypes.Plan{
				Name:   upgradeName,
				Height: int64(haltHeight),
				Info:   "<dummy-info>",
			},
		}
		s.ExecuteGovProposal(msgUpgrade)

		height, err = s.GetFuelSequencerHeight(s.Ctx())
		s.Require().NoError(err, "error fetching height before upgrade")

		// Wait until just before the upgrade
		err = s.WaitUntilSequencerBlock(s.Ctx(), int(haltHeight-1), time.Second*20)
		s.Require().NoError(err)

		// Hold Ethereum so the Sequencer doesn't get too out-of-sync
		s.PauseEthereum()

		// Ensure the nodes have reached the halt height
		time.Sleep(time.Second * 5)

		// Bring down nodes to prepare for upgrade.
		s.Logger().Info("Stopping all sequencer nodes...")
		s.StopAllSequencerNodes()
		s.Logger().Info("Removing all sequencer nodes...")
		s.RemoveAllSequencerNodes()

		// Write new genesis file from state export
		genesis := s.SequencerExportState()
		s.SequencerUnsafeResetAll()
		s.SequencerWriteGenesisFile([]byte(genesis))

		// Resume Ethereum since we're about to resume the Sequencer
		s.UnpauseEthereum()

		// Upgrade version on all nodes and start them back up.
		s.Logger().Info("Starting nodes back up...")
		s.FuelSequencerDockerImageTag = toImageVersion
		s.SequencerRunValidators()

		err = s.WaitForSequencerBlocks(s.Ctx(), int(blocksAfterUpgrade), time.Second*20)
		s.Require().NoError(err, "chain did not produce blocks after upgrade")
	})

	s.Run("Ensure we can perform a large delegation", func() {

		senderAddress := s.SeqKeys[0].AddressSeq
		delegatorKey := s.EthUser.PrivateKey
		delegatorAddress := s.EthUser.AddressHex
		validator1Address := s.SeqKeys[0].ValAddressSeq

		// --------------------------------------- Fund the delegator

		amount, ok := sdkmath.NewIntFromString("10000000000000000000000000000") // 10 bil x 1e18
		s.Require().True(ok)
		delegation := sdk.NewCoin(e2etestsuite.BridgeDenom, amount)
		msgSend := &banktypes.MsgSend{
			FromAddress: senderAddress,
			ToAddress:   delegatorAddress,
			Amount:      sdk.NewCoins(delegation),
		}
		_, err := s.SubmitMsgs(msgSend)
		s.Require().NoError(err)

		// --------------------------------------- Delegate

		// Make sure that there is no pre-existing delegation between the delegator and validator1.
		delegationRaw, err := s.QueryDelegationRaw(s.Ctx(), delegatorAddress, validator1Address)
		s.Require().Nil(delegationRaw)
		s.Require().ErrorContains(
			err,
			fmt.Sprintf("delegation with delegator %s not found for validator %s", delegatorAddress, validator1Address),
		)

		// Generate Authorize event wrapping a MsgDelegate to validator1.
		msgDelegateBz := s.E2ETestSuite.GenerateMsgDelegateBz(delegatorAddress, validator1Address, delegation)
		authorizeData := e2etestsuite.PackAuthorize(msgDelegateBz)
		_, err = s.SendEthTransactionFrom(delegatorKey, e2etestsuite.SequencerInterfaceContractAddress, authorizeData)
		s.Require().NoError(err)

		// Confirm that the delegation went through and is as expected.
		s.PollForDelegationBalance(s.Ctx(), 30, delegatorAddress, validator1Address, delegation)
	})
}
