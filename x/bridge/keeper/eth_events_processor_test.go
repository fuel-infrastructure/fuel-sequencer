package keeper_test

// TODO: move to where we're checking AuthorizeEvent authentication and blocked addresses
//func (s *KeeperTestSuite) TestProcessAuthorizeEvent() {
//	// These accounts correspond to the from and to addresses of the bank.MsgSends to be executed via the AuthorizeEvent
//	fromAcc := sdk.MustAccAddressFromBech32(testtypes.TestFrom3Seq)
//	toAcc, err := s.App.BridgeKeeper.GenerateSequencerAddressFromEthereumAddress(testtypes.TestTo3)
//	s.Require().NoError(err)
//
//	// This is the amount to be funded to the fromAcc
//	amt := sdkmath.NewInt(1000000)
//	coinAmt := sdk.NewCoin("ufuel", amt)
//
//	testCases := []struct {
//		name             string
//		authorizeEvent   *sidecartypes.AuthorizeEvent
//		blockedAddresses map[string]bool
//		expFromBalance   sdkmath.Int
//		expToBalance     sdkmath.Int
//		expErrMsg        string
//	}{
//		{
//			name: "returns error if sender address is blocked",
//			authorizeEvent: &sidecartypes.AuthorizeEvent{
//				Sender: testtypes.TestFrom3,
//				Data:   testutils.MustHexDecodeString(testtypes.TestData4),
//			},
//			blockedAddresses: map[string]bool{
//				fromAcc.String(): true,
//			},
//			expErrMsg: fmt.Sprintf("signer %s is a blocked address", fromAcc.String()),
//		},

//	}
//
//	for _, tc := range testCases {
//		s.Run(tc.name, func() {
//			s.SetupTest()
//
//			// Fund accounts to be used so that we can execute messages
//			s.FundAcc(s.Ctx(), fromAcc, sdk.NewCoins(coinAmt))
//
//			// Authorize bank.MsgSend on Sequencer
//			bridgeParams := &types.Params{
//				AuthorizeMessagesAllowed: []string{"*"},
//			}
//			err := s.App.BridgeKeeper.SetParams(s.Ctx(), *bridgeParams)
//			s.Require().NoError(err)
//
//			processAuthorizeEventCtx := s.Ctx()
//			err = s.App.BridgeKeeper.ProcessAuthorizeEvent(
//				processAuthorizeEventCtx, tc.authorizeEvent, bridgeParams, tc.blockedAddresses,
//			)
//
//			if len(tc.expErrMsg) > 0 {
//				// Confirm that the expected error was raised
//				s.Require().Error(err)
//				s.Require().ErrorContains(err, tc.expErrMsg)
//
//				// Make sure that the balances are as if no message was executed
//				actualFromBalance := s.App.BankKeeper.GetBalance(s.Ctx(), fromAcc, "ufuel")
//				s.Require().Equal(coinAmt.Amount, actualFromBalance.Amount)
//				actualToBalance := s.App.BankKeeper.GetBalance(s.Ctx(), toAcc, "ufuel")
//				s.Require().Equal(sdkmath.ZeroInt(), actualToBalance.Amount)
//				return
//			}
//			s.Require().NoError(err)
//
//			// Confirm that the balances were changed as expected. This indicates that the AuthorizedEvent was executed
//			// successfully
//			actualFromBalance := s.App.BankKeeper.GetBalance(s.Ctx(), fromAcc, "ufuel")
//			s.Require().Equal(tc.expFromBalance, actualFromBalance.Amount)
//			actualToBalance := s.App.BankKeeper.GetBalance(s.Ctx(), toAcc, "ufuel")
//			s.Require().Equal(tc.expToBalance, actualToBalance.Amount)
//		})
//	}
//}
