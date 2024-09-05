package types_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/proto"
	testutiltypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"github.com/stretchr/testify/require"
)

func TestMsgIndex_FromSdkTx(t *testing.T) {

	testMsgIndex := testutiltypes.TestMsgIndex.MsgIndex
	testMsgSend := &banktypes.MsgSend{
		FromAddress: testutiltypes.TestFrom1,
		ToAddress:   testutiltypes.TestTo1,
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 1)),
	}

	testMsgIndexTx := testutiltypes.MustGetTxFromMsgs([]proto.Message{testMsgIndex}, 0)
	testMsgSendTx := testutiltypes.MustGetTxFromMsgs([]proto.Message{testMsgSend}, 0)
	testMixedTx1 := testutiltypes.MustGetTxFromMsgs([]proto.Message{testMsgIndex, testMsgSend}, 0)
	testMixedTx2 := testutiltypes.MustGetTxFromMsgs([]proto.Message{testMsgSend, testMsgIndex}, 0)
	testEmptyTx := testutiltypes.MustGetTxFromMsgs(nil, 0)

	testCases := []struct {
		name        string
		receiver    *types.MsgIndex
		sdkTx       sdk.Tx
		expMsgIndex *types.MsgIndex
		expErrMsg   string
	}{
		{
			name:        "tx with just MsgIndex is correctly parsed",
			receiver:    &types.MsgIndex{},
			sdkTx:       testMsgIndexTx,
			expMsgIndex: testMsgIndex,
		},
		{
			name:      "tx cannot contain non-MsgIndex",
			receiver:  &types.MsgIndex{},
			sdkTx:     testMsgSendTx,
			expErrMsg: fmt.Sprintf("expected msg type URL %s", sdk.MsgTypeURL(&types.MsgIndex{})),
		},
		{
			name:      "tx cannot contain more than one msg (even if MsgIndex is first)",
			receiver:  &types.MsgIndex{},
			sdkTx:     testMixedTx1,
			expErrMsg: "expected 1 msg in MsgIndex raw bytes",
		},
		{
			name:      "tx cannot contain more than one msg (even if MsgIndex is not first)",
			receiver:  &types.MsgIndex{},
			sdkTx:     testMixedTx2,
			expErrMsg: "expected 1 msg in MsgIndex raw bytes",
		},
		{
			name:      "tx cannot be empty",
			receiver:  &types.MsgIndex{},
			sdkTx:     testEmptyTx,
			expErrMsg: "expected 1 msg in MsgIndex raw bytes",
		},
		{
			name:      "receiver cannot be nil",
			receiver:  nil,
			expErrMsg: "expected non-nil MsgIndex receiver",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			err := tc.receiver.FromSdkTx(tc.sdkTx)
			if tc.expErrMsg != "" {
				require.ErrorContains(t, err, tc.expErrMsg)
				return
			}
			require.NoError(t, err)
			require.EqualValues(t, tc.expMsgIndex, tc.receiver)
		})
	}
}
