package utils_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	testutiltypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/utils"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"github.com/stretchr/testify/require"
)

func TestValidRawTxBytesFromAnyMsgs_CorrectEncoding(t *testing.T) {

	sequence := uint64(1) // arbitrary

	msgSupplyDeltaAny, err := codectypes.NewAnyWithValue(testutiltypes.TestMsgSupplyDelta)
	require.NoError(t, err)

	msgIndexAny, err := codectypes.NewAnyWithValue(testutiltypes.TestMsgIndex.MsgIndex)
	require.NoError(t, err)

	msgDepositFromEthereumAny, err := codectypes.NewAnyWithValue(testutiltypes.TestMsgDepositFromEthereum)
	require.NoError(t, err)

	msgSendAny, err := codectypes.NewAnyWithValue(&banktypes.MsgSend{
		FromAddress: testutiltypes.TestFrom1,
		ToAddress:   testutiltypes.TestTo1,
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 1)),
	})
	require.NoError(t, err)

	invalidMsgAny, err := codectypes.NewAnyWithValue(&bridgetypes.Params{})
	require.NoError(t, err)

	testCases := []struct {
		name                    string
		msgAnys                 []*codectypes.Any
		expErrMsgAtFunctionCall string
		expErrMsgAtDecoding     string
	}{
		{
			name:    "MsgSupplyDelta can be used",
			msgAnys: []*codectypes.Any{msgSupplyDeltaAny},
		},
		{
			name:    "MsgIndex can be used",
			msgAnys: []*codectypes.Any{msgIndexAny},
		},
		{
			name:    "MsgDepositFromEthereum can be used",
			msgAnys: []*codectypes.Any{msgDepositFromEthereumAny},
		},
		{
			name:    "MsgSend can be used",
			msgAnys: []*codectypes.Any{msgSendAny},
		},
		{
			name:    "Empty transaction is valid",
			msgAnys: []*codectypes.Any{},
		},
		{
			name:                "Invalid message",
			msgAnys:             []*codectypes.Any{invalidMsgAny},
			expErrMsgAtDecoding: "unable to resolve type URL /fuelsequencer.bridge.Params: tx parse error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			bz, err := utils.ValidRawTxBytesFromAnyMsgs(tc.msgAnys, sequence)
			if tc.expErrMsgAtFunctionCall != "" {
				require.ErrorContains(t, err, tc.expErrMsgAtFunctionCall)
				return
			}
			require.NoError(t, err)

			// Decode to validate
			sdkTx, err := tx.DefaultTxDecoder(testutiltypes.TestCdc)(bz)
			if tc.expErrMsgAtDecoding != "" {
				require.ErrorContains(t, err, tc.expErrMsgAtDecoding)
				return
			}
			require.NoError(t, err)

			// Confirm the transaction's gas consumption is as expected
			gasTx, ok := sdkTx.(baseapp.GasTx)
			require.True(t, ok)
			require.EqualValues(t, utils.InjectedTxGasLimit, gasTx.GetGas())

			// Confirm same number of messages
			require.Equal(t, len(tc.msgAnys), len(sdkTx.GetMsgs()))

			// Confirm messages are unchanged
			for i, msg := range sdkTx.GetMsgs() {
				msgAny, err := codectypes.NewAnyWithValue(msg)
				require.NoError(t, err)
				require.EqualValues(t, tc.msgAnys[i], msgAny)
			}
		})
	}
}

func TestValidRawTxBytesFromAnyMsgs_UniquenessOfSequence(t *testing.T) {

	sequence := uint64(1) // arbitrary

	msgSupplyDeltaAny, err := codectypes.NewAnyWithValue(testutiltypes.TestMsgSupplyDelta)
	require.NoError(t, err)

	msgIndexAny, err := codectypes.NewAnyWithValue(testutiltypes.TestMsgIndex.MsgIndex)
	require.NoError(t, err)

	msgDepositFromEthereumAny, err := codectypes.NewAnyWithValue(testutiltypes.TestMsgDepositFromEthereum)
	require.NoError(t, err)

	msgSendAny, err := codectypes.NewAnyWithValue(&banktypes.MsgSend{
		FromAddress: testutiltypes.TestFrom1,
		ToAddress:   testutiltypes.TestTo1,
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 1)),
	})
	require.NoError(t, err)

	testCases := []struct {
		name    string
		msgAnys []*codectypes.Any
	}{
		{
			name:    "MsgSupplyDelta can be used",
			msgAnys: []*codectypes.Any{msgSupplyDeltaAny},
		},
		{
			name:    "MsgIndex can be used",
			msgAnys: []*codectypes.Any{msgIndexAny},
		},
		{
			name:    "MsgDepositFromEthereum can be used",
			msgAnys: []*codectypes.Any{msgDepositFromEthereumAny},
		},
		{
			name:    "MsgSend can be used",
			msgAnys: []*codectypes.Any{msgSendAny},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			bz, err := utils.ValidRawTxBytesFromAnyMsgs(tc.msgAnys, sequence)
			require.NoError(t, err)
			bzDiffSequence1, err := utils.ValidRawTxBytesFromAnyMsgs(tc.msgAnys, sequence+1)
			require.NoError(t, err)
			bzDiffSequence2, err := utils.ValidRawTxBytesFromAnyMsgs(tc.msgAnys, sequence-1)
			require.NoError(t, err)

			require.NotEqual(t, bz, bzDiffSequence1)
			require.NotEqual(t, bz, bzDiffSequence2)
		})
	}
}
