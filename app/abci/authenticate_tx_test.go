package abci_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *AppTestSuite) TestAuthenticateTx() {
	// Some amounts to populate bank.MsgSend messages
	amt := sdkmath.NewInt(1000000)
	coinAmt := sdk.NewCoin(types.DefaultBridgeDenom, amt)
	coinsAmt := sdk.NewCoins(coinAmt)

	testCases := []struct {
		name             string
		sender           string
		msgs             []sdk.Msg
		blockedAddresses map[string]bool
		expErrMsg        string
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
			blockedAddresses: map[string]bool{},
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
			blockedAddresses: map[string]bool{testtypes.TestFrom1Seq: true},
			expErrMsg:        fmt.Sprintf("signer %s is a blocked address", testtypes.TestFrom1Seq),
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
			blockedAddresses: map[string]bool{},
			expErrMsg:        "could not generate Sequencer address from Ethereum address",
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
			blockedAddresses: map[string]bool{},
			expErrMsg:        "invalid signer",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Execute PrepareProposalHandler
			propHandler := s.GetTestProposalHandler(nil)
			err := propHandler.AuthenticateTx(tc.sender, tc.msgs, tc.blockedAddresses)

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
