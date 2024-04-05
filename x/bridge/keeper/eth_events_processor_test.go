package keeper_test

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
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
		name             string
		authorizeEvent   *sidecartypes.AuthorizeEvent
		blockedAddresses map[string]bool
		expFromBalance   sdkmath.Int
		expToBalance     sdkmath.Int
		expErrMsg        string
	}{
		{
			name: "successfully processes msgs in AuthorizeTx if none error",
			authorizeEvent: &sidecartypes.AuthorizeEvent{
				From: testtypes.TestFrom3,

				// Message decodes two MsgSends of 10 ufuel from testtypes.TestFrom3Seq to testtypes.TestTo3
				Message: testutils.MustHexDecodeString(testtypes.TestMessage4),
			},
			blockedAddresses: map[string]bool{},
			expFromBalance:   sdkmath.NewInt(999980),
			expToBalance:     sdkmath.NewInt(20),
		},
		{
			name: "returns error if sender address is blocked",
			authorizeEvent: &sidecartypes.AuthorizeEvent{
				From:    testtypes.TestFrom3,
				Message: testutils.MustHexDecodeString(testtypes.TestMessage4),
			},
			blockedAddresses: map[string]bool{
				fromAcc.String(): true,
			},
			expErrMsg: fmt.Sprintf("signer %s is a blocked address", fromAcc.String()),
		},
		{
			name: "returns error if AuthorizeTx cannot be deserialized",
			authorizeEvent: &sidecartypes.AuthorizeEvent{
				From:    testtypes.TestFrom3,
				Message: []byte("invalid-message"),
			},
			blockedAddresses: map[string]bool{},
			expErrMsg:        "could not deserialize AuthorizeTx",
		},
		{
			name: "returns error if AuthorizeTx cannot be authenticated",
			authorizeEvent: &sidecartypes.AuthorizeEvent{
				From: testtypes.TestFrom2, // Invalid From to trigger an authentication error
				// Message decodes two MsgSends of 10 ufuel from testtypes.TestFrom3Seq to testtypes.TestTo3
				Message: testutils.MustHexDecodeString(testtypes.TestMessage4),
			},
			blockedAddresses: map[string]bool{},
			expErrMsg:        "could not authenticate AuthorizeTx",
		},
		{
			name: "returns error if some messages cannot be validated",
			authorizeEvent: &sidecartypes.AuthorizeEvent{
				From: testtypes.TestFrom3,

				// Message decodes a MsgWithdrawToEthereum with a zero amount to trigger a failed ValidateBasic.
				Message: testutils.MustHexDecodeString(testtypes.TestMessage5),
			},
			blockedAddresses: map[string]bool{},
			expErrMsg:        "could not validate msg",
		},
		{
			name: "returns error if some messages fail execution",
			authorizeEvent: &sidecartypes.AuthorizeEvent{
				From: testtypes.TestFrom3,

				// Message decodes two MsgSends, one of 10 ufuel and another of 1000000 both from testtypes.TestFrom3Seq
				// to testtypes.TestTo3. The second MsgSend should fail because testtypes.TestFrom3Seq originally should
				// have 1000000 ufuel
				Message: testutils.MustHexDecodeString(testtypes.TestMessage6),
			},
			blockedAddresses: map[string]bool{},
			expErrMsg:        "could not execute msg",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Fund accounts to be used so that we can execute messages
			s.FundAcc(s.Ctx(), fromAcc, sdk.NewCoins(coinAmt))

			// Authorize bank.MsgSend on Sequencer
			bridgeParams := &types.Params{
				AuthorizeMessagesAllowed: []string{"*"},
			}
			err := s.App.BridgeKeeper.SetParams(s.Ctx(), *bridgeParams)
			s.Require().NoError(err)

			processAuthorizeEventCtx := s.Ctx()
			err = s.App.BridgeKeeper.ProcessAuthorizeEvent(
				processAuthorizeEventCtx, tc.authorizeEvent, bridgeParams, tc.blockedAddresses,
			)

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

func (s *KeeperTestSuite) TestExecuteMsg() {
	// These accounts correspond to the from and to addresses of the bank.MsgSend to be executed
	fromAcc := sdk.MustAccAddressFromBech32(testtypes.TestFrom2Seq)
	toAcc := sdk.MustAccAddressFromBech32(testtypes.TestTo3)

	// This is the amount to be funded to the fromAcc
	amt := sdkmath.NewInt(1000000)
	coinAmt := sdk.NewCoin("ufuel", amt)

	testCases := []struct {
		name                  string
		msg                   sdk.Msg
		expFromBalance        sdkmath.Int
		expToBalance          sdkmath.Int
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

			err := s.App.BridgeKeeper.ExecuteMsg(s.Ctx(), tc.msg)

			if len(tc.expErrMsg) > 0 {
				// Confirm that the expected error was raised
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expErrMsg)

				// Make sure that the balances are as expected
				actualFromBalance := s.App.BankKeeper.GetBalance(s.Ctx(), fromAcc, "ufuel")
				s.Require().Equal(tc.expFromBalance, actualFromBalance.Amount)
				actualToBalance := s.App.BankKeeper.GetBalance(s.Ctx(), toAcc, "ufuel")
				s.Require().Equal(tc.expToBalance, actualToBalance.Amount)
				return
			}
			s.Require().NoError(err)

			// Confirm that the balances were changed as expected. This indicates that the AuthorizedEvent was executed
			// successfully
			actualFromBalance := s.App.BankKeeper.GetBalance(s.Ctx(), fromAcc, "ufuel")
			s.Require().Equal(tc.expFromBalance, actualFromBalance.Amount)
			actualToBalance := s.App.BankKeeper.GetBalance(s.Ctx(), toAcc, "ufuel")
			s.Require().Equal(tc.expToBalance, actualToBalance.Amount)
		})
	}
}

func (s *KeeperTestSuite) TestProcessEthereumEvents_SendToSequencerEvent() {

	blockTime, _ := time.Parse(time.DateOnly, "2024-01-01")
	govAddr := s.App.AccountKeeper.GetModuleAddress(govtypes.ModuleName)

	// These accounts correspond to the from and to addresses of the bank.MsgSend to be executed via the AuthorizeEvent
	fromAccOne, _ := types.GenerateSequencerAddressFromEthereumAddress(testtypes.TestFrom3)
	toAccOne := sdk.MustAccAddressFromBech32(testtypes.TestTo3)

	fromAccTwo, _ := types.GenerateSequencerAddressFromEthereumAddress(testtypes.TestFrom2)

	testCases := []struct {
		name           string
		ethEventsTx    *types.EthEventsTx
		toAcc          *sdk.AccAddress
		fromAcc        *sdk.AccAddress
		expFromBalance sdkmath.Int
		expToBalance   sdkmath.Int
		expSupplyDelta *types.SupplyDeltaInfo
		expGovBal      sdkmath.Int
	}{
		{
			name: "successful - send to sequencer - mint to To address",
			ethEventsTx: &types.EthEventsTx{
				Events: []*sidecartypes.Event{
					testtypes.TestEvent1,
				},
				AdvanceSequencer: true,
				NewEthereumBlock: true,
				BlockNumber:      sdkmath.OneInt(),
			},
			fromAcc:        &fromAccOne,
			toAcc:          &toAccOne,
			expFromBalance: sdkmath.NewInt(0),
			expToBalance:   sdkmath.NewInt(102),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-102),
			},
			expGovBal: sdkmath.ZeroInt(),
		},
		{
			name: "successful - send to sequencer - mint to To address twice",
			ethEventsTx: &types.EthEventsTx{
				Events: []*sidecartypes.Event{
					testtypes.TestEvent1,
					testtypes.TestEvent1,
				},
				AdvanceSequencer: true,
				NewEthereumBlock: true,
				BlockNumber:      sdkmath.OneInt(),
			},
			fromAcc:        &fromAccOne,
			toAcc:          &toAccOne,
			expFromBalance: sdkmath.NewInt(0),
			expToBalance:   sdkmath.NewInt(204),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-204),
			},
			expGovBal: sdkmath.ZeroInt(),
		},
		{
			name: "successful - send to sequencer - mint to From address",
			ethEventsTx: &types.EthEventsTx{
				Events: []*sidecartypes.Event{
					testtypes.TestEvent3,
				},
				AdvanceSequencer: true,
				NewEthereumBlock: true,
				BlockNumber:      sdkmath.OneInt(),
			},
			fromAcc:        &fromAccTwo,
			toAcc:          nil,
			expFromBalance: sdkmath.NewInt(101),
			expToBalance:   sdkmath.NewInt(0),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-101),
			},
			expGovBal: sdkmath.ZeroInt(),
		},
		{
			name: "successful - send to sequencer - mint to From address twice",
			ethEventsTx: &types.EthEventsTx{
				Events: []*sidecartypes.Event{
					testtypes.TestEvent3,
					testtypes.TestEvent3,
				},
				AdvanceSequencer: true,
				NewEthereumBlock: true,
				BlockNumber:      sdkmath.OneInt(),
			},
			fromAcc:        &fromAccTwo,
			toAcc:          nil,
			expFromBalance: sdkmath.NewInt(202),
			expToBalance:   sdkmath.NewInt(0),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-202),
			},
			expGovBal: sdkmath.ZeroInt(),
		},
		{
			name: "failure - send to sequencer - duration failed to parse - mint to governance",
			ethEventsTx: &types.EthEventsTx{
				Events: []*sidecartypes.Event{
					testtypes.TestEvent4,
				},
				AdvanceSequencer: true,
				NewEthereumBlock: true,
				BlockNumber:      sdkmath.OneInt(),
			},
			fromAcc:        &fromAccOne,
			toAcc:          &toAccOne,
			expFromBalance: sdkmath.NewInt(0),
			expToBalance:   sdkmath.NewInt(0),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-102),
			},
			expGovBal: sdkmath.NewInt(102),
		},
		{
			name: "failure - send to sequencer - from address failed to parse - mint to governance",
			ethEventsTx: &types.EthEventsTx{
				Events: []*sidecartypes.Event{
					testtypes.TestEvent5,
				},
				AdvanceSequencer: true,
				NewEthereumBlock: true,
				BlockNumber:      sdkmath.OneInt(),
			},
			fromAcc:        &fromAccOne,
			toAcc:          &toAccOne,
			expFromBalance: sdkmath.NewInt(0),
			expToBalance:   sdkmath.NewInt(0),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-102),
			},
			expGovBal: sdkmath.NewInt(102),
		},
		{
			name: "failure - send to sequencer - bad vesting duration - mint to governance",
			ethEventsTx: &types.EthEventsTx{
				Events: []*sidecartypes.Event{
					testtypes.TestEvent6,
				},
				AdvanceSequencer: true,
				NewEthereumBlock: true,
				BlockNumber:      sdkmath.OneInt(),
			},
			fromAcc:        &fromAccOne,
			toAcc:          nil,
			expFromBalance: sdkmath.NewInt(0),
			expToBalance:   sdkmath.NewInt(0),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-102),
			},
			expGovBal: sdkmath.NewInt(102),
		},
		{
			name: "failure - send to sequencer - bad to bech32 address - mint to governance",
			ethEventsTx: &types.EthEventsTx{
				Events: []*sidecartypes.Event{
					testtypes.TestEvent7,
				},
				AdvanceSequencer: true,
				NewEthereumBlock: true,
				BlockNumber:      sdkmath.OneInt(),
			},
			fromAcc:        &fromAccOne,
			toAcc:          &toAccOne,
			expFromBalance: sdkmath.NewInt(0),
			expToBalance:   sdkmath.NewInt(0),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-102),
			},
			expGovBal: sdkmath.NewInt(102),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Set LastEthereumBlockSynced to EthEventsTxs' block height
			s.App.BridgeKeeper.SetLastEthereumBlockSynced(s.Ctx(), tc.ethEventsTx.BlockNumber)

			// Authorize bank.MsgSend on Sequencer
			err := s.App.BridgeKeeper.SetParams(s.Ctx(), types.Params{
				BridgeDenom:              "ufuel",
				AuthorizeMessagesAllowed: []string{"*"},
				VestingStartTime:         blockTime,
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

			// Confirm that the balances were changed as specified. This indicates that the AuthorizedEvents were
			// executed successfully
			if tc.fromAcc != nil {
				actualFromBalance := s.App.BankKeeper.GetBalance(s.Ctx(), *tc.fromAcc, "ufuel")
				s.Require().Equal(tc.expFromBalance, actualFromBalance.Amount)
			}
			if tc.toAcc != nil {
				actualToBalance := s.App.BankKeeper.GetBalance(s.Ctx(), *tc.toAcc, "ufuel")
				s.Require().Equal(tc.expToBalance, actualToBalance.Amount)
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

func (s *KeeperTestSuite) TestProcessEthereumEventsSendToSequencerEvent_AmountParseFailure() {
	blockTime, _ := time.Parse(time.RFC3339, "2024-01-01T00:00:00Z")
	govAddr := s.App.AccountKeeper.GetModuleAddress(govtypes.ModuleName)

	ethEventsTx := &types.EthEventsTx{
		Events: []*sidecartypes.Event{
			testtypes.TestEvent8,
		},
		AdvanceSequencer: true,
		NewEthereumBlock: true,
		BlockNumber:      sdkmath.OneInt(),
	}
	s.SetupTest() // Reset the test suite

	// Set initial conditions: LastEthereumBlockSynced and Params
	s.App.BridgeKeeper.SetLastEthereumBlockSynced(s.Ctx(), ethEventsTx.BlockNumber)

	// Authorize bank.MsgSend on Sequencer
	err := s.App.BridgeKeeper.SetParams(s.Ctx(), types.Params{
		BridgeDenom:              "ufuel",
		AuthorizeMessagesAllowed: []string{"*"},
		VestingStartTime:         blockTime,
	})
	s.Require().NoError(err)

	// Set the malformed EthEventsTx
	s.App.BridgeKeeper.SetEthEventsTx(s.Ctx(), *ethEventsTx)
	_, found := s.App.BridgeKeeper.GetEthEventsTx(s.Ctx(), ethEventsTx.BlockNumber.Uint64())
	s.Require().True(found)

	// Trigger the processing of the Ethereum events
	s.Require().Panics(func() {
		s.App.BridgeKeeper.ProcessEthereumEvents(s.Ctx())
	})

	// Verify the governance address balance is as expected
	actualGovBal := s.App.BankKeeper.GetBalance(s.Ctx(), govAddr, "ufuel")
	s.Require().Equal(sdkmath.ZeroInt(), actualGovBal.Amount)
}
