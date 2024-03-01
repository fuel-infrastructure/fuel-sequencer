package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/fuel-infrastructure/fuel-sequencer/app/apptesting"
	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	queryClient types.QueryClient
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()

	s.queryClient = types.NewQueryClient(s.QueryHelper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
