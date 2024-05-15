package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

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
//			name: "successfully processes msgs in AuthorizeTx if none error",
//			authorizeEvent: &sidecartypes.AuthorizeEvent{
//				Sender: testtypes.TestFrom3,
//
//				// Data decodes two MsgSends of 10 ufuel from testtypes.TestFrom3Seq to testtypes.TestTo3
//				Data: testutils.MustHexDecodeString(testtypes.TestData4),
//			},
//			blockedAddresses: map[string]bool{},
//			expFromBalance:   sdkmath.NewInt(999980),
//			expToBalance:     sdkmath.NewInt(20),
//		},
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
//		{
//			name: "returns error if AuthorizeTx cannot be deserialized",
//			authorizeEvent: &sidecartypes.AuthorizeEvent{
//				Sender: testtypes.TestFrom3,
//				Data:   []byte("invalid-data"),
//			},
//			blockedAddresses: map[string]bool{},
//			expErrMsg:        "could not deserialize AuthorizeTx",
//		},
//		{
//			name: "returns error if AuthorizeTx cannot be authenticated",
//			authorizeEvent: &sidecartypes.AuthorizeEvent{
//				Sender: testtypes.TestFrom2, // Invalid Sender to trigger an authentication error
//
//				// Data decodes two MsgSends of 10 ufuel from testtypes.TestFrom3Seq to testtypes.TestTo3
//				Data: testutils.MustHexDecodeString(testtypes.TestData4),
//			},
//			blockedAddresses: map[string]bool{},
//			expErrMsg:        "could not authenticate AuthorizeTx",
//		},
//		{
//			name: "returns error if some messages cannot be validated",
//			authorizeEvent: &sidecartypes.AuthorizeEvent{
//				Sender: testtypes.TestFrom3,
//
//				// Data decodes a MsgWithdrawToEthereum with a zero amount to trigger a failed ValidateBasic.
//				Data: testutils.MustHexDecodeString(testtypes.TestData5),
//			},
//			blockedAddresses: map[string]bool{},
//			expErrMsg:        "could not validate msg",
//		},
//		{
//			name: "returns error if some messages fail execution",
//			authorizeEvent: &sidecartypes.AuthorizeEvent{
//				Sender: testtypes.TestFrom3,
//
//				// Data decodes two MsgSends, one of 10 ufuel and another of 1000000 both from testtypes.TestFrom3Seq
//				// to testtypes.TestTo3. The second MsgSend should fail because testtypes.TestFrom3Seq originally should
//				// have 1000000 ufuel
//				Data: testutils.MustHexDecodeString(testtypes.TestData6),
//			},
//			blockedAddresses: map[string]bool{},
//			expErrMsg:        "could not execute msg",
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
		blockedAddresses   map[string]bool
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
			blockedAddresses:   map[string]bool{},
			authorizedMessages: []string{"*"},
		},
		{
			name:   "errors if signer address is blocked",
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
			blockedAddresses:   map[string]bool{testtypes.TestFrom1Seq: true},
			authorizedMessages: []string{"*"},
			expErrMsg:          fmt.Sprintf("signer %s is a blocked address", testtypes.TestFrom1Seq),
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
			blockedAddresses:   map[string]bool{},
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
			blockedAddresses:   map[string]bool{},
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
			blockedAddresses:   map[string]bool{},
			authorizedMessages: []string{sdk.MsgTypeURL(&banktypes.MsgSend{}), sdk.MsgTypeURL(&types.MsgSupplyDelta{})},
			expErrMsg:          "invalid signer",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Authorize required messages on Sequencer
			bridgeParams := &types.Params{
				AuthorizeMessagesAllowed: tc.authorizedMessages,
			}
			err := s.App.BridgeKeeper.SetParams(s.Ctx(), *bridgeParams)
			s.Require().NoError(err)

			err = s.App.BridgeKeeper.AuthenticateTx(tc.sender, tc.msgs, bridgeParams, tc.blockedAddresses)

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
