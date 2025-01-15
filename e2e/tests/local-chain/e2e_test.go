package authorize_transactions_test

import (
	"testing"

	e2etestsuite "github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	"github.com/stretchr/testify/suite"
)

type LocalChainTestSuite struct {
	e2etestsuite.E2ETestSuite
}

func TestAuthorizeTransactionsTestSuite(t *testing.T) {
	suite.Run(t, new(LocalChainTestSuite))
}
