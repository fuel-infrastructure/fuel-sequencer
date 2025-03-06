package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/keeper"

	"github.com/fuel-infrastructure/fuel-sequencer/app/apptesting"
	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	queryClient types.QueryClient
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()

	s.queryClient = types.NewQueryClient(s.QueryHelper)
}

func (s *KeeperTestSuite) GetMsgServer() types.MsgServer {
	return keeper.NewMsgServerImpl(s.App.ReportsKeeper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
