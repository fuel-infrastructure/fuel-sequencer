package authorize_fixtures_test

import (
	"testing"

	e2etestsuite "github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	"github.com/stretchr/testify/suite"
)

type AuthorizeFixturesTestSuite struct {
	e2etestsuite.E2ETestSuite
}

func TestAuthorizeFixturesTestSuite(t *testing.T) {
	suite.Run(t, new(AuthorizeFixturesTestSuite))
}

// SetupTest modifies the genesis file as required by AuthorizeFixturesTestSuite
func (s *AuthorizeFixturesTestSuite) SetupTest() {
	s.E2ETestSuite.SetupTest()
}
