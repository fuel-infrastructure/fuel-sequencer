package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/fuel-infrastructure/fuel-sequencer/app/apptesting"
	testkeeper "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

var expectedAuthority = authtypes.NewModuleAddress(govtypes.ModuleName).String()

func TestNewKeeper(t *testing.T) {
	k, _ := testkeeper.BondKeeper(t)
	require.NotNil(t, k)
	require.Equal(t, expectedAuthority, k.GetAuthority())
}

func TestKeeper_GetAuthority(t *testing.T) {
	k, _ := testkeeper.BondKeeper(t)
	require.Equal(t, expectedAuthority, k.GetAuthority())
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
