package keeper_test

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/keeper"
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

func (s *KeeperTestSuite) TestDepositFromEthereum() {

	blockTime, _ := time.Parse(time.DateOnly, "2024-01-01")
	govAddr := s.App.AccountKeeper.GetModuleAddress(govtypes.ModuleName)

	// These accounts correspond to the from and to addresses of the DepositEvent
	fromAccOne, _ := s.App.BridgeKeeper.GenerateSequencerAddressFromEthereumAddress(testtypes.TestFrom3)
	toAccOne, err := s.App.BridgeKeeper.GenerateSequencerAddressFromEthereumAddress(testtypes.TestTo3)
	s.Require().NoError(err)

	fromAccTwo, _ := s.App.BridgeKeeper.GenerateSequencerAddressFromEthereumAddress(testtypes.TestFrom2)

	testCases := []struct {
		name           string
		msgs           []*types.MsgDepositFromEthereum
		toAcc          *sdk.AccAddress
		fromAcc        *sdk.AccAddress
		expFromBalance sdkmath.Int
		expToBalance   sdkmath.Int
		expSupplyDelta *types.SupplyDeltaInfo
		expGovBal      sdkmath.Int
		isToEthOwned   bool
		isFromEthOwned bool
	}{
		{
			name: "successful - deposit - mint to Recipient address",
			msgs: []*types.MsgDepositFromEthereum{
				testtypes.TestEvent1Msg,
			},
			fromAcc:        &fromAccOne,
			toAcc:          &toAccOne,
			expFromBalance: sdkmath.NewInt(0),
			expToBalance:   sdkmath.NewInt(102),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-102),
			},
			expGovBal:      sdkmath.ZeroInt(),
			isToEthOwned:   false,
			isFromEthOwned: false,
		},
		{
			name: "successful - deposit - mint to Recipient address twice",
			msgs: []*types.MsgDepositFromEthereum{
				testtypes.TestEvent1Msg,
				testtypes.TestEvent1Msg,
			},
			fromAcc:        &fromAccOne,
			toAcc:          &toAccOne,
			expFromBalance: sdkmath.NewInt(0),
			expToBalance:   sdkmath.NewInt(204),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-204),
			},
			expGovBal:      sdkmath.ZeroInt(),
			isToEthOwned:   false,
			isFromEthOwned: false,
		},
		{
			name: "successful - deposit - mint to address Recipient == Depositor",
			msgs: []*types.MsgDepositFromEthereum{
				testtypes.TestEvent10Msg,
			},
			fromAcc:        &fromAccOne,
			toAcc:          &fromAccOne,
			expFromBalance: sdkmath.NewInt(100), // Same balance Depositor == Recipient
			expToBalance:   sdkmath.NewInt(100),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-100),
			},
			expGovBal:      sdkmath.ZeroInt(),
			isToEthOwned:   true,
			isFromEthOwned: true,
		},
		{
			name: "successful - deposit - mint to address Recipient == Sequencer(Depositor)",
			msgs: []*types.MsgDepositFromEthereum{
				testtypes.TestEvent11Msg,
			},
			fromAcc:        &fromAccOne,
			toAcc:          &fromAccOne,
			expFromBalance: sdkmath.NewInt(100), // Same balance Depositor == Recipient
			expToBalance:   sdkmath.NewInt(100),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-100),
			},
			expGovBal:      sdkmath.ZeroInt(),
			isToEthOwned:   true,
			isFromEthOwned: true,
		},
		{
			name: "successful - deposit - mint to Depositor address",
			msgs: []*types.MsgDepositFromEthereum{
				testtypes.TestEvent3Msg,
			},
			fromAcc:        &fromAccTwo,
			toAcc:          nil,
			expFromBalance: sdkmath.NewInt(101),
			expToBalance:   sdkmath.NewInt(0),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-101),
			},
			expGovBal:      sdkmath.ZeroInt(),
			isToEthOwned:   false,
			isFromEthOwned: true,
		},
		{
			name: "successful - deposit - mint to Depositor address twice",
			msgs: []*types.MsgDepositFromEthereum{
				testtypes.TestEvent3Msg,
				testtypes.TestEvent3Msg,
			},
			fromAcc:        &fromAccTwo,
			toAcc:          nil,
			expFromBalance: sdkmath.NewInt(202),
			expToBalance:   sdkmath.NewInt(0),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-202),
			},
			expGovBal:      sdkmath.ZeroInt(),
			isToEthOwned:   false,
			isFromEthOwned: true,
		},
		{
			name: "failure - deposit - lockup failed to parse - mint to governance",
			msgs: []*types.MsgDepositFromEthereum{
				testtypes.TestEvent4Msg,
			},
			fromAcc:        &fromAccOne,
			toAcc:          &toAccOne,
			expFromBalance: sdkmath.NewInt(0),
			expToBalance:   sdkmath.NewInt(0),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-102),
			},
			expGovBal:      sdkmath.NewInt(102),
			isToEthOwned:   false,
			isFromEthOwned: false,
		},
		{
			name: "failure - deposit - from address failed to parse - mint to governance",
			msgs: []*types.MsgDepositFromEthereum{
				testtypes.TestEvent5Msg,
			},
			fromAcc:        &fromAccOne,
			toAcc:          &toAccOne,
			expFromBalance: sdkmath.NewInt(0),
			expToBalance:   sdkmath.NewInt(0),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-102),
			},
			expGovBal:      sdkmath.NewInt(102),
			isToEthOwned:   false,
			isFromEthOwned: false,
		},
		{
			name: "failure - deposit - bad vesting duration - mint to governance",
			msgs: []*types.MsgDepositFromEthereum{
				testtypes.TestEvent6Msg,
			},
			fromAcc:        &fromAccOne,
			toAcc:          nil,
			expFromBalance: sdkmath.NewInt(0),
			expToBalance:   sdkmath.NewInt(0),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-102),
			},
			expGovBal:      sdkmath.NewInt(102),
			isToEthOwned:   false,
			isFromEthOwned: false,
		},
		{
			name: "failure - deposit - bad to bech32 address - mint to governance",
			msgs: []*types.MsgDepositFromEthereum{
				testtypes.TestEvent7Msg,
			},
			fromAcc:        &fromAccOne,
			toAcc:          &toAccOne,
			expFromBalance: sdkmath.NewInt(0),
			expToBalance:   sdkmath.NewInt(0),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-102),
			},
			expGovBal:      sdkmath.NewInt(102),
			isToEthOwned:   false,
			isFromEthOwned: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Get the message server
			msgServer := keeper.NewMsgServerImpl(s.App.BridgeKeeper)

			// Authorize bank.MsgSend on Sequencer
			err := s.App.BridgeKeeper.SetParams(s.Ctx(), types.Params{
				BridgeDenom:              "ufuel",
				AuthorizeMessagesAllowed: []string{"*"},
				VestingStartTime:         blockTime,
			})
			s.Require().NoError(err)

			for _, msg := range tc.msgs {
				_, err = msgServer.DepositFromEthereum(s.Ctx(), msg)
				s.Require().NoError(err)
			}

			// Confirm that the balances were changed as specified. This indicates that the AuthorizedEvents were
			// executed successfully
			if tc.fromAcc != nil {
				actualFromBalance := s.App.BankKeeper.GetBalance(s.Ctx(), *tc.fromAcc, "ufuel")
				s.Require().Equal(tc.expFromBalance, actualFromBalance.Amount)

				// Verify if the account is ETH owned
				account := s.App.AccountKeeper.GetAccount(s.Ctx(), *tc.fromAcc)
				_, ok := account.(types.EthOwnedAccountI)
				s.Require().Equal(tc.isFromEthOwned, ok)
			}
			if tc.toAcc != nil {
				actualToBalance := s.App.BankKeeper.GetBalance(s.Ctx(), *tc.toAcc, "ufuel")
				s.Require().Equal(tc.expToBalance, actualToBalance.Amount)

				// Verify if the account is ETH owned
				account := s.App.AccountKeeper.GetAccount(s.Ctx(), *tc.toAcc)
				_, ok := account.(types.EthOwnedAccountI)
				s.Require().Equal(tc.isToEthOwned, ok)
			}

			// Verify the supply delta info was updated
			if tc.expSupplyDelta != nil {
				actualSupplyDelta := s.App.BridgeKeeper.MustGetSupplyDeltaInfo(s.Ctx())
				s.Require().Equal(tc.expSupplyDelta.Offset, actualSupplyDelta.Offset)
			}

			// Verify the governance address balance is as expected
			actualGovBal := s.App.BankKeeper.GetBalance(s.Ctx(), govAddr, "ufuel")
			s.Require().Equal(tc.expGovBal, actualGovBal.Amount)
		})
	}
}

func (s *KeeperTestSuite) TestDepositFromEthereum_AmountParseFailure() {
	blockTime, _ := time.Parse(time.RFC3339, "2024-01-01T00:00:00Z")
	govAddr := s.App.AccountKeeper.GetModuleAddress(govtypes.ModuleName)

	msg := testtypes.TestEvent8Msg
	s.SetupTest() // Reset the test suite

	// Authorize bank.MsgSend on Sequencer
	err := s.App.BridgeKeeper.SetParams(s.Ctx(), types.Params{
		BridgeDenom:              "ufuel",
		AuthorizeMessagesAllowed: []string{"*"},
		VestingStartTime:         blockTime,
	})
	s.Require().NoError(err)

	// Get the message server
	msgServer := keeper.NewMsgServerImpl(s.App.BridgeKeeper)

	// Trigger the processing of the Ethereum events
	s.Require().Panics(func() {
		_, _ = msgServer.DepositFromEthereum(s.Ctx(), msg)
	})

	// Verify the governance address balance is as expected
	actualGovBal := s.App.BankKeeper.GetBalance(s.Ctx(), govAddr, "ufuel")
	s.Require().Equal(sdkmath.ZeroInt(), actualGovBal.Amount)
}
