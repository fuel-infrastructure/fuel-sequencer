package features_and_optimisations_test

import (
	"math/big"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/fuel-infrastructure/fuel-sequencer/app/upgrades/features_and_optimisations"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
)

const (
	haltHeightDelta    = uint64(25) // will propose upgrade this many blocks in the future; must be > voting period
	blocksAfterUpgrade = uint64(10) // will wait for this many blocks after the upgrade
	upgradeName        = features_and_optimisations.UpgradeName
	fromImageVersion   = "b543d8d" // this image needs to exist for this test to run
	toImageVersion     = "d25152c" // this image needs to exist for this test to run
)

type UpgradesTestSuite struct {
	testsuite.E2ETestSuite
}

func TestUpgradesTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradesTestSuite))
}

func (s *UpgradesTestSuite) SetupTest() {

	s.FuelSequencerDockerImageTag = fromImageVersion

	s.E2ETestSuite.SetupTest()
}

func (s *UpgradesTestSuite) TestUpgrade() {

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

		// Resume Ethereum since we're about to resume the Sequencer
		s.UnpauseEthereum()

		// Upgrade version on all nodes and start them back up.
		s.Logger().Info("Starting nodes back up...")
		s.FuelSequencerDockerImageTag = toImageVersion
		s.RunSequencerValidators()

		err = s.WaitForSequencerBlocks(s.Ctx(), int(blocksAfterUpgrade), time.Second*20)
		s.Require().NoError(err, "chain did not produce blocks after upgrade")
	})

	s.Run("Check that bridge module params were migrated", func() {

		// If the query works it's enough evidence that the params were obtained successfully.
		_ = s.QueryBridgeParams(s.Ctx())
	})

	s.Run("Check that Ethereum events get picked up by the Sidecar and processed by the Sequencer", func() {

		// Try getting height (RPC).
		ethHeight, err := s.GetEthereumHeight(s.Ctx())
		s.Require().NoError(err)
		s.Require().Greater(ethHeight, uint64(1))

		// Try generating some events via a transaction (RPC) - via deposit.
		depositAmount := big.NewInt(200)
		depositTxReceipt := s.DepositTokenToSequencer(depositAmount)

		// Generate a Transfer
		sendAmount := int64(10)
		from := s.EthKeys[0]
		to := s.EthKeys[1]
		transfer := testsuite.PackTransfer(to.Address, big.NewInt(sendAmount))
		msgSendBz := s.E2ETestSuite.GenerateMsgSendBz(
			from.AddressHex, to.AddressHex,
			sdk.NewCoins(sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewInt(sendAmount))),
		)

		sendTxReceipt, err := s.SendEthTransactionToSequencerInterfaceContract(transfer)
		s.Require().NoError(err)

		// --------------------------------------- Ensure Sidecar got the new Events

		// Ensure deposit event is at the expected height.
		depositEvents, err := s.PollForSidecarBlockEvents(s.Ctx(), time.Second*20, int(depositTxReceipt.BlockNumber.Int64()))
		s.Require().NoError(err)
		s.Require().Len(depositEvents, 1)
		s.Require().Equal(sidecartypes.DepositEventName, depositEvents[0].EventType)

		// Check deposit event data is as expected
		var depositEventData sidecartypes.DepositEvent
		err = depositEventData.Unmarshal(depositEvents[0].Data)
		s.Require().NoError(err)
		s.Require().Equal(depositEventData, sidecartypes.DepositEvent{
			Depositor: s.EthKeys[0].AddressHex,
			Recipient: s.EthKeys[0].AddressHex, // sender == recipient unless otherwise specified
			Amount:    depositAmount.String(),
			Lockup:    "0", // the deposit initiated from Ethereum has no lockup
		})

		// Ensure authorize event is at the expected height.
		authorizeEvents, err := s.PollForSidecarBlockEvents(
			s.Ctx(), time.Second*20, int(sendTxReceipt.BlockNumber.Int64()),
		)
		s.Require().NoError(err)
		s.Require().Len(authorizeEvents, 1)
		s.Require().Equal(sidecartypes.AuthorizeEventName, authorizeEvents[0].EventType)

		// Check authorize event data is as expected
		var authorizeEventData sidecartypes.AuthorizeEvent
		err = authorizeEventData.Unmarshal(authorizeEvents[0].Data)
		s.Require().NoError(err)
		s.Require().Equal(authorizeEventData, sidecartypes.AuthorizeEvent{
			Sender: from.AddressHex,
			Data:   msgSendBz,
		})

		// --------------------------------------- Wait until Sequencer has processed the events

		s.PollForLastEthereumBlockSynced(s.Ctx(), 20, sendTxReceipt.BlockNumber.Uint64())

		// --------------------------------------- Ensure deposit and transfer went through

		depositAmountU64 := depositAmount.Uint64()
		sendAmountU64 := uint64(sendAmount)

		bal0, err := s.QueryBalance(s.Ctx(), s.EthKeys[0].AddressSeq, testsuite.BridgeDenom)
		s.Require().NoError(err)
		s.Require().Equal(bal0.Balance.Amount.Uint64(), depositAmountU64-sendAmountU64)

		bal1, err := s.QueryBalance(s.Ctx(), s.EthKeys[1].AddressSeq, testsuite.BridgeDenom)
		s.Require().NoError(err)
		s.Require().Equal(bal1.Balance.Amount.Uint64(), sendAmountU64)
	})

	s.Run("Grant and Revoke from Ethereum to Claim Rewards from the Sequencer on behalf of a granter", func() {

		granter, grantee := s.EthKeys[0], s.SeqKeys[0]

		// No grant yet
		grants := s.QueryGranterGrants(s.Ctx(), granter.AddressSeq)
		s.Require().Empty(grants)

		// Grant an authorisation from Ethereum, and wait for the grant to be processed
		grantData := testsuite.PackGrantClaimRewards(grantee.AddressEth, 0)
		txReceipt, err := s.SendEthTransactionToSequencerInterfaceContract(grantData)
		s.Require().NoError(err)
		s.PollForLastEthereumBlockSynced(s.Ctx(), 10, txReceipt.BlockNumber.Uint64())

		// Grant exists now
		grants = s.QueryGranterGrants(s.Ctx(), granter.AddressSeq)
		s.Require().Len(grants, 1)
		grant := grants[0]
		s.Require().Equal(grantee.AddressSeq, grant.Grantee)
		s.Require().Equal(granter.AddressSeq, grant.Granter)

		// Check the grant's authorization
		var authorization authz.Authorization
		s.Require().NoError(testsuite.TestCdc.UnpackAny(grant.Authorization, &authorization))
		genericAuthz, ok := authorization.(*authz.GenericAuthorization)
		s.Require().True(ok)
		s.Require().Equal(cdctypes.MsgTypeURL(&distrtypes.MsgWithdrawDelegatorReward{}), genericAuthz.Msg)

		// Revoke the authorisation from Ethereum, and wait for the revoke to be processed
		revokeData := testsuite.PackRevokeClaimRewards(grantee.AddressEth)
		txReceipt, err = s.SendEthTransactionToSequencerInterfaceContract(revokeData)
		s.Require().NoError(err)
		s.PollForLastEthereumBlockSynced(s.Ctx(), 10, txReceipt.BlockNumber.Uint64())

		// No grant anymore
		grants = s.QueryGranterGrants(s.Ctx(), granter.AddressSeq)
		s.Require().Empty(grants)
	})
}
