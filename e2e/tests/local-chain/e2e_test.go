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

func (s *LocalChainTestSuite) SetupTest() {
	// Use only two mnemonics
	e2etestsuite.MNEMONICS = e2etestsuite.MNEMONICS[:2]

	// Only Sequencer nodes needed
	s.SequencerOnly = true

	s.E2ETestSuite.SetupTest()
}
