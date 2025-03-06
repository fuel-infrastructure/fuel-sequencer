package basic_test

import (
	"testing"

	e2etestsuite "github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	"github.com/stretchr/testify/suite"
)

type SpecialMsgsTestSuite struct {
	e2etestsuite.E2ETestSuite
}

func TestSpecialMsgsTestSuite(t *testing.T) {
	suite.Run(t, new(SpecialMsgsTestSuite))
}
