package types_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	testutiltypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func TestMsgSupplyDelta_FromSdkTx(t *testing.T) {

	testMsgSupplyDelta := &types.MsgSupplyDelta{Authority: testutiltypes.TestGovernanceAddress}
	testMsgSend := &banktypes.MsgSend{
		FromAddress: testutiltypes.TestFrom1,
		ToAddress:   testutiltypes.TestTo1,
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 1)),
	}

	sequence := uint64(1) // arbitrary
	testMsgSupplyDeltaTx := testutiltypes.MustGetTxFromMsgs([]proto.Message{testMsgSupplyDelta}, sequence)
	testMsgSendTx := testutiltypes.MustGetTxFromMsgs([]proto.Message{testMsgSend}, sequence)
	testMixedTx1 := testutiltypes.MustGetTxFromMsgs([]proto.Message{testMsgSupplyDelta, testMsgSend}, sequence)
	testMixedTx2 := testutiltypes.MustGetTxFromMsgs([]proto.Message{testMsgSend, testMsgSupplyDelta}, sequence)
	testEmptyTx := testutiltypes.MustGetTxFromMsgs(nil, sequence)

	testCases := []struct {
		name              string
		receiver          *types.MsgSupplyDelta
		sdkTx             sdk.Tx
		expMsgSupplyDelta *types.MsgSupplyDelta
		expErrMsg         string
	}{
		{
			name:              "tx with just MsgSupplyDelta is correctly parsed",
			receiver:          &types.MsgSupplyDelta{},
			sdkTx:             testMsgSupplyDeltaTx,
			expMsgSupplyDelta: testMsgSupplyDelta,
		},
		{
			name:      "tx cannot contain non-MsgSupplyDelta",
			receiver:  &types.MsgSupplyDelta{},
			sdkTx:     testMsgSendTx,
			expErrMsg: fmt.Sprintf("expected msg type URL %s", sdk.MsgTypeURL(&types.MsgSupplyDelta{})),
		},
		{
			name:      "tx cannot contain more than one msg (even if MsgSupplyDelta is first)",
			receiver:  &types.MsgSupplyDelta{},
			sdkTx:     testMixedTx1,
			expErrMsg: "expected 1 msg in MsgSupplyDelta raw bytes",
		},
		{
			name:      "tx cannot contain more than one msg (even if MsgSupplyDelta is not first)",
			receiver:  &types.MsgSupplyDelta{},
			sdkTx:     testMixedTx2,
			expErrMsg: "expected 1 msg in MsgSupplyDelta raw bytes",
		},
		{
			name:      "tx cannot be empty",
			receiver:  &types.MsgSupplyDelta{},
			sdkTx:     testEmptyTx,
			expErrMsg: "expected 1 msg in MsgSupplyDelta raw bytes",
		},
		{
			name:      "receiver cannot be nil",
			receiver:  nil,
			expErrMsg: "expected non-nil MsgSupplyDelta receiver",
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
			require.EqualValues(t, tc.expMsgSupplyDelta, tc.receiver)
		})
	}
}
