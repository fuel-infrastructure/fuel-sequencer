package mint_test

import (
	"testing"

	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/suite"

	"github.com/fuel-infrastructure/fuel-sequencer/app/apptesting"
)

type MintModuleTestSuite struct {
	apptesting.KeeperTestHelper

	queryClient minttypes.QueryClient
}

func (s *MintModuleTestSuite) SetupTest() {
	s.Setup()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(MintModuleTestSuite))
}
