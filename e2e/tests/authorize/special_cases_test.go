package authorize_test

// TODO: Restore this test, by sending a deposit and delegate combined message, with the sequencer params configured to allow only 1 message.

// func (s *AuthorizeTestSuite) TestAuthorizeEvents_AuthorizeWithTooManyMessagesIsSkipped() {
// 	s.Run("An AuthorizeEvent with more messages than MaxAuthorizeMessages is skipped", func() {
// 		senderAddress := s.EthKeys[0].AddressHex
// 		receiverAddress := s.EthKeys[1].AddressHex

// 		// Generate and send a deposit and delegate combined message.
// 		sendAmount := big.NewInt(10)
// 		msgDepositAndDelegate := testsuite.PackDepositAndDelegate(sendAmount, validatorAddress)
// 		_, err := s.SendEthTransactionToSequencerInterfaceContract(msgDepositAndDelegate)
// 		s.Require().NoError(err)

// 		// TODO: Generate and send a deposit and delegate combined message.
// 		// sendAmount := big.NewInt(10)
// 		// msgDepositAndDelegate := testsuite.PackDepositAndDelegate(sendAmount, validatorAddress)
// 		// _, err := s.SendEthTransactionToSequencerInterfaceContract(msgDepositAndDelegate)
// 		// s.Require().NoError(err)

// 		// Send Authorize tx
// 		authorizeData := testsuite.PackAuthorize(msgSendBz)
// 		_, err := s.SendEthTransactionToSequencerInterfaceContract(authorizeData)
// 		s.Require().NoError(err)

// 		// Make sure that the Sequencer skips the event
// 		re := regexp.MustCompile(
// 			"skipping event; failed to encode event as raw tx bytes with err: authorize event has too many messages",
// 		)
// 		s.Require().Eventually(func() bool {
// 			return len(s.FindSequencerLogs(re)) > 0
// 		}, time.Minute, time.Second)
// 	})
// }
