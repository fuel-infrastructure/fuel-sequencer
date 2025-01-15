package authorize_transactions_test

import (
	"time"
)

func (s *LocalChainTestSuite) TestAuthorizedTransactions_MsgSend() {
	_ = s.WaitForSequencerBlocks(s.Ctx(), 99999999, time.Hour)
}
