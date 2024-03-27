package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

			// Confirm that the balances were changed as specified. This indicates that the AuthorizedEvents were
			// executed successfully
			actualFromBalance := s.App.BankKeeper.GetBalance(s.Ctx(), fromAcc, "ufuel")
			s.Require().Equal(tc.expFromBalance, actualFromBalance.Amount)
			actualToBalance := s.App.BankKeeper.GetBalance(s.Ctx(), toAcc, "ufuel")
			s.Require().Equal(tc.expToBalance, actualToBalance.Amount)

			// TODO (Vitaly): Add Assertions for SendToSequencer events
		})
	}
}
