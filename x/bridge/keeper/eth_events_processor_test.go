package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/proto"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	testutils "github.com/fuel-infrastructure/fuel-sequencer/testutil"
	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	bridgekeeper "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *KeeperTestSuite) TestProcessEthereumEvents_AuthorizeEvent() {
	// These accounts correspond to the from and to addresses of the bank.MsgSend to be executed via the AuthorizeEvent
	fromAcc := sdk.MustAccAddressFromBech32(testtypes.TestFrom3Seq)
	toAcc := sdk.MustAccAddressFromBech32(testtypes.TestTo3)

	// This is the amount to be funded to the fromAcc
	amt := sdkmath.NewInt(1000000)
	coinAmt := sdk.NewCoin("ufuel", amt)

	testCases := []struct {
		name           string
		ethEventsTx    *types.EthEventsTx
		expFromBalance sdkmath.Int
		expToBalance   sdkmath.Int
	}{
		{
			name: "successfully processes AuthorizeEvents if none error",
			ethEventsTx: &types.EthEventsTx{
				Events:           []*sidecartypes.Event{testtypes.TestEvent2, testtypes.TestEvent2},
				AdvanceSequencer: true,
				NewEthereumBlock: true,
				BlockNumber:      sdkmath.OneInt(),
			},
			expFromBalance: sdkmath.NewInt(999980),
			expToBalance:   sdkmath.NewInt(20),
		},
		{
			name: "successfully processes valid AuthorizeEvents if some cannot be unmarshalled",
			ethEventsTx: &types.EthEventsTx{
				Events: []*sidecartypes.Event{
					{EventType: "invalid-event", Data: []byte("invalid bytes")}, // event with unrecognized type
					testtypes.TestEvent2, // valid AuthorizeEvent
				},
				AdvanceSequencer: true,
				NewEthereumBlock: true,
				BlockNumber:      sdkmath.OneInt(),
			},
			expFromBalance: sdkmath.NewInt(999990),
			expToBalance:   sdkmath.NewInt(10),
		},
		{
			name: "successfully processes valid AuthorizeEvents if some error while executing",
			ethEventsTx: &types.EthEventsTx{
				Events: []*sidecartypes.Event{
					// Event that sends 10 ufuel from a sequencer address that has no funds. We expect this to fail,
					// demonstrating that other events still get processed successfully
					testutils.MustGetSidecarEventFromParsedEvent(&sidecartypes.AuthorizeEvent{
						From:    testtypes.TestFrom2,
						Message: testutils.MustHexDecodeString(testtypes.TestMessage2),
					}),
					testtypes.TestEvent2, // valid AuthorizeEvent
				},
				AdvanceSequencer: true,
				NewEthereumBlock: true,
				BlockNumber:      sdkmath.OneInt(),
			},
			expFromBalance: sdkmath.NewInt(999990),
			expToBalance:   sdkmath.NewInt(10),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Fund accounts to be used so that we can execute messages
			s.FundAcc(s.Ctx(), fromAcc, sdk.NewCoins(coinAmt))

			// Set LastEthereumBlockSynced to EthEventsTxs' block height
			s.App.BridgeKeeper.SetLastEthereumBlockSynced(s.Ctx(), tc.ethEventsTx.BlockNumber)

			// Authorize bank.MsgSend on Sequencer
			err := s.App.BridgeKeeper.SetParams(s.Ctx(), types.Params{
				AuthorizeMessagesAllowed: []string{"*"},
			})
			s.Require().NoError(err)

			// Set EthEventsTx
			s.App.BridgeKeeper.SetEthEventsTx(s.Ctx(), *tc.ethEventsTx)
			_, found := s.App.BridgeKeeper.GetEthEventsTx(s.Ctx(), tc.ethEventsTx.BlockNumber.Uint64())
			s.Require().True(found)

			s.App.BridgeKeeper.ProcessEthereumEvents(s.Ctx())

			// Make sure that EthEventsTx has been removed
			_, found = s.App.BridgeKeeper.GetEthEventsTx(s.Ctx(), tc.ethEventsTx.BlockNumber.Uint64())
			s.Require().False(found)

			// Confirm that the balances were changed as expected. This indicates that the AuthorizedEvents were
			// executed successfully
			actualFromBalance := s.App.BankKeeper.GetBalance(s.Ctx(), fromAcc, "ufuel")
			s.Require().Equal(tc.expFromBalance, actualFromBalance.Amount)
			actualToBalance := s.App.BankKeeper.GetBalance(s.Ctx(), toAcc, "ufuel")
			s.Require().Equal(tc.expToBalance, actualToBalance.Amount)
		})
	}
}

func (s *KeeperTestSuite) TestProcessAuthorizeEvent() {
	// These accounts correspond to the from and to addresses of the bank.MsgSends to be executed via the AuthorizeEvent
	fromAcc := sdk.MustAccAddressFromBech32(testtypes.TestFrom3Seq)
	toAcc := sdk.MustAccAddressFromBech32(testtypes.TestTo3)

	// This is the amount to be funded to the fromAcc
	amt := sdkmath.NewInt(1000000)
	coinAmt := sdk.NewCoin("ufuel", amt)

	testCases := []struct {
		name           string
		authorizeEvent *sidecartypes.AuthorizeEvent
		expFromBalance sdkmath.Int
		expToBalance   sdkmath.Int
		expErrMsg      string
	}{
		{
			name: "successfully processes msgs in AuthorizeTx if none error",
			authorizeEvent: &sidecartypes.AuthorizeEvent{
				From: testtypes.TestFrom3,
				Message: testutils.MustHexDecodeString(
					"0aae010a1c2f636f736d6f732e62616e6b2e763162657461312e4d736753656e64128d010a486675656c73657175656e" +
						"636572317373796d66356a796b61383967736a633966616d657a76326c6373656437756c646671327a38746b6d76" +
						"6739713365746a327171757a7261356e12346675656c73657175656e636572313633727376363574343839337432" +
						"727a35726d646139736c79376c67646c71326a677233366d1a0b0a05756675656c120231300aae010a1c2f636f73" +
						"6d6f732e62616e6b2e763162657461312e4d736753656e64128d010a486675656c73657175656e63657231737379" +
						"6d66356a796b61383967736a633966616d657a76326c6373656437756c646671327a38746b6d766739713365746a" +
						"327171757a7261356e12346675656c73657175656e636572313633727376363574343839337432727a35726d6461" +
						"39736c79376c67646c71326a677233366d1a0b0a05756675656c12023130",
				), // Message decodes two MsgSends of 10 ufuel from testtypes.TestFrom3Seq to testtypes.TestTo3
			},
			expFromBalance: sdkmath.NewInt(999980),
			expToBalance:   sdkmath.NewInt(20),
		},
		{
			name: "returns error if AuthorizeTx cannot be deserialized",
			authorizeEvent: &sidecartypes.AuthorizeEvent{
				From:    testtypes.TestFrom3,
				Message: []byte("invalid-message"),
			},
			expErrMsg: "could not deserialize AuthorizeTx",
		},
		{
			name: "returns error if AuthorizeTx cannot be authenticated",
			authorizeEvent: &sidecartypes.AuthorizeEvent{
				From: testtypes.TestFrom2, // Invalid From to trigger an authentication error
				Message: testutils.MustHexDecodeString(
					"0aae010a1c2f636f736d6f732e62616e6b2e763162657461312e4d736753656e64128d010a486675656c73657175656e" +
						"636572317373796d66356a796b61383967736a633966616d657a76326c6373656437756c646671327a38746b6d76" +
						"6739713365746a327171757a7261356e12346675656c73657175656e636572313633727376363574343839337432" +
						"727a35726d646139736c79376c67646c71326a677233366d1a0b0a05756675656c120231300aae010a1c2f636f73" +
						"6d6f732e62616e6b2e763162657461312e4d736753656e64128d010a486675656c73657175656e63657231737379" +
						"6d66356a796b61383967736a633966616d657a76326c6373656437756c646671327a38746b6d766739713365746a" +
						"327171757a7261356e12346675656c73657175656e636572313633727376363574343839337432727a35726d6461" +
						"39736c79376c67646c71326a677233366d1a0b0a05756675656c12023130",
				), // Message decodes two MsgSends of 10 ufuel from testtypes.TestFrom3Seq to testtypes.TestTo3
			},
			expErrMsg: "could not authenticate AuthorizeTx",
		},
		{
			name: "returns error if some messages cannot be validated",
			authorizeEvent: &sidecartypes.AuthorizeEvent{
				From: testtypes.TestFrom3,
				Message: testutils.MustHexDecodeString(
					"0abc010a2b2f6675656c73657175656e6365722e6272696467652e4d73675769746864726177546f457468657265756" +
						"d128c010a486675656c73657175656e636572317373796d66356a796b61383967736a633966616d657a76326c63" +
						"73656437756c646671327a38746b6d766739713365746a327171757a7261356e12346675656c73657175656e636" +
						"572313633727376363574343839337432727a35726d646139736c79376c67646c71326a677233366d1a0a0a0575" +
						"6675656c120130",
				), // Message decodes a MsgWithdrawToEthereum with a zero amount to trigger a failed ValidateBasic.
			},
			expErrMsg: "could not validate msg",
		},
		{
			name: "returns error if some messages fail execution",
			authorizeEvent: &sidecartypes.AuthorizeEvent{
				From: testtypes.TestFrom3,
				Message: testutils.MustHexDecodeString(
					"0aae010a1c2f636f736d6f732e62616e6b2e763162657461312e4d736753656e64128d010a486675656c73657175656e" +
						"636572317373796d66356a796b61383967736a633966616d657a76326c6373656437756c646671327a38746b6d76" +
						"6739713365746a327171757a7261356e12346675656c73657175656e636572313633727376363574343839337432" +
						"727a35726d646139736c79376c67646c71326a677233366d1a0b0a05756675656c120231300ab4010a1c2f636f73" +
						"6d6f732e62616e6b2e763162657461312e4d736753656e641293010a486675656c73657175656e63657231737379" +
						"6d66356a796b61383967736a633966616d657a76326c6373656437756c646671327a38746b6d766739713365746a" +
						"327171757a7261356e12346675656c73657175656e636572313633727376363574343839337432727a35726d6461" +
						"39736c79376c67646c71326a677233366d1a110a05756675656c12083130303030303030",
				), // Message decodes two MsgSends, one of 10 ufuel and another of 1000000 both from
				// testtypes.TestFrom3Seq to testtypes.TestTo3. The second MsgSend should fail because
				// testtypes.TestFrom3Seq originally should have 1000000 ufuel
			},
			expErrMsg: "could not execute msg",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Fund accounts to be used so that we can execute messages
			s.FundAcc(s.Ctx(), fromAcc, sdk.NewCoins(coinAmt))

			// Authorize bank.MsgSend on Sequencer
			err := s.App.BridgeKeeper.SetParams(s.Ctx(), types.Params{
				AuthorizeMessagesAllowed: []string{"*"},
			})
			s.Require().NoError(err)

			processAuthorizeEventCtx := s.Ctx()
			err = s.App.BridgeKeeper.ProcessAuthorizeEvent(processAuthorizeEventCtx, tc.authorizeEvent)

			if len(tc.expErrMsg) > 0 {
				// Confirm that the expected error was raised
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expErrMsg)

				// Make sure that the balances are as if no message was executed
				actualFromBalance := s.App.BankKeeper.GetBalance(s.Ctx(), fromAcc, "ufuel")
				s.Require().Equal(coinAmt.Amount, actualFromBalance.Amount)
				actualToBalance := s.App.BankKeeper.GetBalance(s.Ctx(), toAcc, "ufuel")
				s.Require().Equal(sdkmath.ZeroInt(), actualToBalance.Amount)
				return
			}
			s.Require().NoError(err)

			// Confirm that the balances were changed as expected. This indicates that the AuthorizedEvent was executed
			// successfully
			actualFromBalance := s.App.BankKeeper.GetBalance(s.Ctx(), fromAcc, "ufuel")
			s.Require().Equal(tc.expFromBalance, actualFromBalance.Amount)
			actualToBalance := s.App.BankKeeper.GetBalance(s.Ctx(), toAcc, "ufuel")
			s.Require().Equal(tc.expToBalance, actualToBalance.Amount)

			// Check events emitted
			s.AssertEventEmitted(processAuthorizeEventCtx, proto.MessageName(&types.EventAuthorizedTxExecuted{}), 1)
		})
	}
}

func (s *KeeperTestSuite) TestAuthenticateTx() {
	// Some amounts to populate bank.MsgSend messages
	amt := sdkmath.NewInt(1000000)
	coinAmt := sdk.NewCoin("ufuel", amt)
	coinsAmt := sdk.NewCoins(coinAmt)

	testCases := []struct {
		name               string
		sender             string
		msgs               []sdk.Msg
		authorizedMessages []string
		expErrMsg          string
	}{
		{
			name:   "authenticates valid msgs successfully",
			sender: testtypes.TestFrom1,
			msgs: []sdk.Msg{
				&banktypes.MsgSend{
					FromAddress: testtypes.TestFrom1Seq,
					ToAddress:   testtypes.TestTo3,
					Amount:      coinsAmt,
				},
				&banktypes.MsgSend{
					FromAddress: testtypes.TestFrom1Seq,
					ToAddress:   testtypes.TestTo2,
					Amount:      coinsAmt,
				},
			},
			authorizedMessages: []string{"*"},
		},
		{
			name:   "errors if sender cannot be mapped to its Sequencer address",
			sender: "invalid-eth-address",
			msgs: []sdk.Msg{
				&banktypes.MsgSend{
					FromAddress: testtypes.TestFrom1Seq,
					ToAddress:   testtypes.TestTo3,
					Amount:      coinsAmt,
				},
				&banktypes.MsgSend{
					FromAddress: testtypes.TestFrom1Seq,
					ToAddress:   testtypes.TestTo2,
					Amount:      coinsAmt,
				},
			},
			authorizedMessages: []string{"*"},
			expErrMsg:          "could not generate Sequencer address from Ethereum address",
		},
		{
			name:   "errors if one of the messages is not authorized",
			sender: testtypes.TestFrom1,
			msgs: []sdk.Msg{
				&banktypes.MsgSend{
					FromAddress: testtypes.TestFrom1Seq,
					ToAddress:   testtypes.TestTo3,
					Amount:      coinsAmt,
				},
				&types.MsgSupplyDelta{
					Authority: testtypes.TestGovernanceAddress,
				},
			},
			authorizedMessages: []string{sdk.MsgTypeURL(&banktypes.MsgSend{})},
			expErrMsg:          "message not authorized on Sequencer",
		},
		{
			name:   "errors if one of the messages' signer is not as expected",
			sender: testtypes.TestFrom1,
			msgs: []sdk.Msg{
				&banktypes.MsgSend{
					FromAddress: testtypes.TestFrom1Seq,
					ToAddress:   testtypes.TestTo3,
					Amount:      coinsAmt,
				},
				&types.MsgSupplyDelta{
					Authority: testtypes.TestGovernanceAddress, // Message signer not equivalent to testtypes.TestFrom1
				},
			},
			authorizedMessages: []string{sdk.MsgTypeURL(&banktypes.MsgSend{}), sdk.MsgTypeURL(&types.MsgSupplyDelta{})},
			expErrMsg:          "invalid signer",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Authorize required messages on Sequencer
			err := s.App.BridgeKeeper.SetParams(s.Ctx(), types.Params{
				AuthorizeMessagesAllowed: tc.authorizedMessages,
			})
			s.Require().NoError(err)

			err = s.App.BridgeKeeper.AuthenticateTx(tc.sender, tc.msgs, tc.authorizedMessages)

			if len(tc.expErrMsg) > 0 {
				// Confirm that the expected error was raised
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expErrMsg)
				return
			}

			s.Require().NoError(err)
		})
	}
}

func (s *KeeperTestSuite) TestExecuteMsg() {
	// These accounts correspond to the from and to addresses of the bank.MsgSend to be executed
	fromAcc := sdk.MustAccAddressFromBech32(testtypes.TestFrom2Seq)
	toAcc := sdk.MustAccAddressFromBech32(testtypes.TestTo3)

	// This is the amount to be funded to the fromAcc
	amt := sdkmath.NewInt(1000000)
	coinAmt := sdk.NewCoin("ufuel", amt)

	expMsgSendResponse, err := codectypes.NewAnyWithValue(&banktypes.MsgSendResponse{})
	s.Require().NoError(err)

	testCases := []struct {
		name                  string
		msg                   sdk.Msg
		expFromBalance        sdkmath.Int
		expToBalance          sdkmath.Int
		expResponse           *codectypes.Any
		resetMsgServiceRouter bool
		expErrMsg             string
	}{
		{
			name: "successfully executes recognized msg",
			msg: &banktypes.MsgSend{
				FromAddress: testtypes.TestFrom2Seq,
				ToAddress:   testtypes.TestTo3,
				Amount:      sdk.NewCoins(sdk.NewCoin("ufuel", sdkmath.NewInt(10))),
			},
			expFromBalance:        sdkmath.NewInt(999990),
			expToBalance:          sdkmath.NewInt(10),
			expResponse:           expMsgSendResponse,
			resetMsgServiceRouter: false,
		},
		{
			name: "returns error if unrecognized msg",
			msg: &banktypes.MsgSend{
				FromAddress: testtypes.TestFrom2Seq,
				ToAddress:   testtypes.TestTo3,
				Amount:      sdk.NewCoins(sdk.NewCoin("ufuel", sdkmath.NewInt(10))),
			},
			expFromBalance:        amt,
			expToBalance:          sdkmath.ZeroInt(),
			expResponse:           nil,
			resetMsgServiceRouter: true, // No message will be registered
			expErrMsg:             "invalid MsgHandler route",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			if tc.resetMsgServiceRouter {
				s.App.BridgeKeeper = bridgekeeper.SetRouter(s.App.BridgeKeeper, baseapp.NewMsgServiceRouter())
			}

			// Fund accounts to be used so that we can execute messages
			s.FundAcc(s.Ctx(), fromAcc, sdk.NewCoins(coinAmt))

			res, err := s.App.BridgeKeeper.ExecuteMsg(s.Ctx(), tc.msg)

			if len(tc.expErrMsg) > 0 {
				// Confirm that the expected error was raised
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expErrMsg)
				s.Require().Nil(res)

				// Make sure that the balances are as expected
				actualFromBalance := s.App.BankKeeper.GetBalance(s.Ctx(), fromAcc, "ufuel")
				s.Require().Equal(tc.expFromBalance, actualFromBalance.Amount)
				actualToBalance := s.App.BankKeeper.GetBalance(s.Ctx(), toAcc, "ufuel")
				s.Require().Equal(tc.expToBalance, actualToBalance.Amount)
				return
			}
			s.Require().NoError(err)

			// Check that the message response was as expected
			s.Require().Equal(tc.expResponse, res)

			// Confirm that the balances were changed as expected. This indicates that the AuthorizedEvent was executed
			// successfully
			actualFromBalance := s.App.BankKeeper.GetBalance(s.Ctx(), fromAcc, "ufuel")
			s.Require().Equal(tc.expFromBalance, actualFromBalance.Amount)
			actualToBalance := s.App.BankKeeper.GetBalance(s.Ctx(), toAcc, "ufuel")
			s.Require().Equal(tc.expToBalance, actualToBalance.Amount)
		})
	}
}
