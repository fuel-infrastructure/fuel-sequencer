package upgrades_test

import (
	"testing"
	"time"

	upgradetypes "cosmossdk.io/x/upgrade/types"
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

	// Ensure the nodes have reached the halt height
	time.Sleep(time.Second * 5)

	// Bring down nodes to prepare for upgrade.
	s.Logger().Info("Stopping all sequencer nodes...")
	s.StopAllSequencerNodes()
	s.Logger().Info("Removing all sequencer nodes...")
	s.RemoveAllSequencerNodes()

	// Upgrade version on all nodes and start them back up.
	s.FuelSequencerDockerImageTag = toImageVersion
	s.RunFuelSequencerValidators()

	err = s.WaitForSequencerBlocks(s.Ctx(), int(blocksAfterUpgrade), time.Second*20)
	s.Require().NoError(err, "chain did not produce blocks after upgrade")
}
