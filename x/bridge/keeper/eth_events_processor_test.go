package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	testutils "github.com/fuel-infrastructure/fuel-sequencer/testutil"
	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *KeeperTestSuite) TestProcessEthereumEvents() {
	// These accounts correspond to the from and to addresses of the bank.MsgSend to be executed via the AuthorizeEvent
	fromAcc := sdk.MustAccAddressFromBech32(testtypes.TestFrom3Seq)
	toAcc := sdk.MustAccAddressFromBech32(testtypes.TestTo3)

	testCases := []struct {
		name           string
		ethEventsTx    *types.EthEventsTx
		expFromBalance sdkmath.Int
		expToBalance   sdkmath.Int
	}{
		{
			name:           "successfully processes Ethereum events if none error",
			ethEventsTx:    testtypes.TestEthEventsTx, // Contains 2 SendToSequencer and 1 Authorize events
			expFromBalance: sdkmath.NewInt(999990),
			expToBalance:   sdkmath.NewInt(10),
		},
		{
			name: "successfully processes valid Ethereum events if some cannot be unmarshalled",
			ethEventsTx: &types.EthEventsTx{
				Events: []*sidecartypes.Event{
					testtypes.TestEvent1, // valid SendToSequencerEvent
					{EventType: "invalid-event", Data: []byte("invalid bytes")}, // event with unrecognized type
					testtypes.TestEvent2, // valid AuthorizeEvent
					testtypes.TestEvent3, // valid SendToSequencerEvent
				},
				AdvanceSequencer: true,
				NewEthereumBlock: true,
				BlockNumber:      sdkmath.OneInt(),
			},
			expFromBalance: sdkmath.NewInt(999990),
			expToBalance:   sdkmath.NewInt(10),
		},
		{
			name: "successfully processes valid Ethereum events if some error while executing",
			ethEventsTx: &types.EthEventsTx{
				Events: []*sidecartypes.Event{
					testtypes.TestEvent1, // valid SendToSequencerEvent

					// Event that sends 10 ufuel from a sequencer address that has no funds. We expect this to fail,
					// demonstrating that other events still get processed successfully
					testutils.MustGetSidecarEventFromParsedEvent(&sidecartypes.AuthorizeEvent{
						From:    testtypes.TestFrom2,
						Message: testutils.MustHexDecodeString(testtypes.TestMessage2),
					}),

					// TODO (Vitaly): Might need to add a SendToSequencerEvent that errors during execution

					testtypes.TestEvent2, // valid AuthorizeEvent
					testtypes.TestEvent3, // valid SendToSequencerEvent
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
			amt := sdkmath.NewInt(1000000)
			s.FundAcc(s.Ctx(), fromAcc, sdk.NewCoins(sdk.NewCoin("ufuel", amt)))

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

			// TODO (Vitaly): Add Assertions for SendToSequencer events
		})
	}
}

func (s *KeeperTestSuite) TestProcessAuthorizeEvent() {
	// These accounts correspond to the from and to addresses of the bank.MsgSends to be executed via the AuthorizeEvent
	fromAcc := sdk.MustAccAddressFromBech32(testtypes.TestFrom3Seq)
	toAcc := sdk.MustAccAddressFromBech32(testtypes.TestTo3)

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
		// TODO: Cases below have valid text but we just need to fix the params
		//{
		//	name:           "returns error if AuthorizeTx cannot be deserialized",
		//	ethEventsTx:    testtypes.TestEthEventsTx, // Contains 2 SendToSequencer and 1 Authorize events
		//	expFromBalance: sdkmath.NewInt(999990),
		//	expToBalance:   sdkmath.NewInt(10),
		//},
		//{
		//	name:           "returns error if AuthorizeTx cannot be authenticated",
		//	ethEventsTx:    testtypes.TestEthEventsTx, // Contains 2 SendToSequencer and 1 Authorize events
		//	expFromBalance: sdkmath.NewInt(999990),
		//	expToBalance:   sdkmath.NewInt(10),
		//},
		//{
		//	name:           "returns error if some messages cannot be validated",
		//	ethEventsTx:    testtypes.TestEthEventsTx, // Contains 2 SendToSequencer and 1 Authorize events
		//	expFromBalance: sdkmath.NewInt(999990),
		//	expToBalance:   sdkmath.NewInt(10),
		//},
		//{
		//	name:           "returns error if some messages fail execution",
		//	ethEventsTx:    testtypes.TestEthEventsTx, // Contains 2 SendToSequencer and 1 Authorize events
		//	expFromBalance: sdkmath.NewInt(999990),
		//	expToBalance:   sdkmath.NewInt(10),
		//},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Fund accounts to be used so that we can execute messages
			amt := sdkmath.NewInt(1000000)
			coinAmt := sdk.NewCoin("ufuel", amt)
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
				s.Require().Equal(coinAmt, actualFromBalance.Amount)
				actualToBalance := s.App.BankKeeper.GetBalance(s.Ctx(), toAcc, "ufuel")
				s.Require().Equal(sdk.NewCoin("ufuel", sdkmath.ZeroInt()), actualToBalance.Amount)
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
