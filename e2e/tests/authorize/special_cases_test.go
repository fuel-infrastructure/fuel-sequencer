package authorize_test

import (
	"regexp"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
)

func (s *AuthorizeTestSuite) TestAuthorizeEvents_InvalidDataDoesNotCauseHalt() {
	s.Run("Invalid data from Ethereum causes Sequencer to skip an invalid authorize event", func() {

		// Generate Authorize event wrapping invalid data.
		invalidBz := []byte("some invalid data")
		authorizeData := testsuite.PackAuthorize(invalidBz)
		_, err := s.SendEthTransactionToSequencerInterfaceContract(authorizeData)
		s.Require().NoError(err)

		// Check that the Sequencer skips the event
		re := regexp.MustCompile("skipping event; failed to encode event as raw tx bytes with err")
		s.Require().Eventually(func() bool {
			return len(s.FindSequencerLogs(re)) > 0
		}, time.Minute, time.Second)
	})
}

func (s *AuthorizeTestSuite) TestAuthorizeEvents_AuthorizeWithTooManyMessagesIsSkipped() {
	s.Run("An AuthorizeEvent with more messages than MaxAuthorizeMessages is skipped", func() {
		senderAddress := s.EthKeys[0].AddressHex
		receiverAddress := s.EthKeys[1].AddressHex

		// Generate Authorize tx with 100 MsgSends.
		sendAmount, ok := sdkmath.NewIntFromString("10")
		s.Require().True(ok)
		sendCoin := sdk.NewCoin(testsuite.BridgeDenom, sendAmount)
		sendCoins := sdk.NewCoins(sendCoin)
		msgSend := banktypes.MsgSend{
			FromAddress: senderAddress,
			ToAddress:   receiverAddress,
			Amount:      sendCoins,
		}
		msgSendBz := s.E2ETestSuite.GenerateNMsgsBz(&msgSend, 100)

		// Send Authorize tx
		authorizeData := testsuite.PackAuthorize(msgSendBz)
		_, err := s.SendEthTransactionToSequencerInterfaceContract(authorizeData)
		s.Require().NoError(err)

		// Make sure that the Sequencer skips the event
		re := regexp.MustCompile(
			"skipping event; failed to encode event as raw tx bytes with err: authorize event has too many messages",
		)
		s.Require().Eventually(func() bool {
			return len(s.FindSequencerLogs(re)) > 0
		}, time.Minute, time.Second)
	})
}
