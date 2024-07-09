package mint_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/fuel-infrastructure/fuel-sequencer/app/apptesting"
)

type MintModuleTestSuite struct {
	apptesting.KeeperTestHelper
}

func (s *MintModuleTestSuite) SetupTest() {
	s.Setup()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(MintModuleTestSuite))
}
