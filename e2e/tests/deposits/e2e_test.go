package deposits_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	e2etestsuite "github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
)

type DepositsTestSuite struct {
	e2etestsuite.E2ETestSuite
}

func TestDepositsTestSuite(t *testing.T) {
	suite.Run(t, new(DepositsTestSuite))
}
