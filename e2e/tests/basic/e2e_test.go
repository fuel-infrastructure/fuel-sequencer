package basic_test

import (
	"testing"

	e2etestsuite "github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	"github.com/stretchr/testify/suite"
)

type BasicTestSuite struct {
	e2etestsuite.E2ETestSuite
}

func (s *BasicTestSuite) SetupTest() {
	if testing.Short() {
		s.T().Skip()
	}

	// TODO: s.Reset()
}

func TestBasicTestSuite(t *testing.T) {
	suite.Run(t, new(BasicTestSuite))
}
