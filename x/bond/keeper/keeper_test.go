package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/fuel-infrastructure/fuel-sequencer/app/apptesting"
	testkeeper "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

func TestNewKeeper(t *testing.T) {
	k, _ := testkeeper.BondKeeper(t)
	require.NotNil(t, k)
	require.Equal(t, "fuelsequencer1w8rk2mk84wytpxx7ld63kaqpkhmd39m05xlgt4", k.GetAuthority())
}

func TestKeeper_GetAuthority(t *testing.T) {
	k, _ := testkeeper.BondKeeper(t)
	require.Equal(t, "fuelsequencer1w8rk2mk84wytpxx7ld63kaqpkhmd39m05xlgt4", k.GetAuthority())
}

func TestKeeper_Logger(t *testing.T) {
	k, _ := testkeeper.BondKeeper(t)
	require.NotNil(t, k.Logger())
}

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	queryClient types.QueryClient
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()

	s.queryClient = types.NewQueryClient(s.QueryHelper)
}

func (s *KeeperTestSuite) GetMsgServer() types.MsgServer {
	return keeper.NewMsgServerImpl(s.App.BondKeeper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
