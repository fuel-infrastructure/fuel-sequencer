package keeper_test

import (
	"testing"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/keeper"
	"github.com/stretchr/testify/suite"

	"github.com/fuel-infrastructure/fuel-sequencer/app/apptesting"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
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
	return keeper.NewMsgServerImpl(s.App.BridgeKeeper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
