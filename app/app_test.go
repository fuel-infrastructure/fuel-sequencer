package app_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/fuel-infrastructure/fuel-sequencer/app/apptesting"
)

type AppTestSuite struct {
	apptesting.KeeperTestHelper
}

func (s *AppTestSuite) SetupTest() {
	s.Setup()
}

func TestAppTestSuite(t *testing.T) {
	suite.Run(t, new(AppTestSuite))
}
