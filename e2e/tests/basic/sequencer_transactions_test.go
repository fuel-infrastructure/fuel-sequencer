package basic_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	banktypes "cosmossdk.io/x/bank/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	sequencingtypes "github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"
)

func (s *BasicTestSuite) TestSequencerTransactions_NativeTxGetsRejectedIfTooLarge() {
	s.Run("Large transactions submitted to the Sequencer get rejected if they exceed limit", func() {

		// Set SequencerTxMaxBytes to a very low number so that we make sure that any submitted tx will get rejected.
		sequencerTxMaxBytes := uint64(1)
		sequencingParams := s.QuerySequencingParams(s.Ctx())
		sequencingParams.SequencerTxMaxBytes = sequencerTxMaxBytes
		msgUpdateParams := &sequencingtypes.MsgUpdateParams{
			Authority: s.GetGovernanceAddress(),
			Params:    *sequencingParams,
		}
		s.ExecuteGovProposal(msgUpdateParams)

		// Check that Sequencer tx max bytes was updated correctly
		sequencingParams = s.QuerySequencingParams(s.Ctx())
		s.Require().EqualValues(sequencerTxMaxBytes, sequencingParams.SequencerTxMaxBytes)

		// Generate a MsgSend natively on the Sequencer
		msgSend := &banktypes.MsgSend{
			FromAddress: s.SeqKeys[0].AddressSeq,
			ToAddress:   s.SeqKeys[1].AddressSeq,
			Amount:      sdk.NewCoins(sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewInt(10))),
		}
		resp, err := s.SubmitMsgs(msgSend)
		s.Require().NoError(err)

		// This is the error message we expect.
		expectedErr := fmt.Sprintf("transaction is too large")
		s.Require().Contains(resp.RawLog, expectedErr)
	})
}
