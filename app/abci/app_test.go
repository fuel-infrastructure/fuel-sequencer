package abci_test

import (
	"testing"

	"github.com/fuel-infrastructure/fuel-sequencer/app/abci"
	"github.com/fuel-infrastructure/fuel-sequencer/app/apptesting"
	sidecartestutil "github.com/fuel-infrastructure/fuel-sequencer/sidecar/testutil"
	"github.com/stretchr/testify/suite"
)

type AppTestSuite struct {
	apptesting.KeeperTestHelper
}

// GetTestProposalHandler simply returns a ProposalHandler with the specified sidecar client mock
func (s *AppTestSuite) GetTestProposalHandler(
	sidecarClientMock *sidecartestutil.MockAppSidecarClient,
) *abci.FuelSequencerProposalHandler {
	return abci.NewFuelSequencerProposalHandler(
		s.App.Logger(), s.App.StakingKeeper, s.App, sidecarClientMock, s.App.BridgeKeeper,
	)
}

func (s *AppTestSuite) SetupTest() {
	s.Setup()
}

func TestAppTestSuite(t *testing.T) {
	suite.Run(t, new(AppTestSuite))
}
