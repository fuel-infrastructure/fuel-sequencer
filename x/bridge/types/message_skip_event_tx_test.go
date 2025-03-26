package types_test

import (
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/stretchr/testify/require"

	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/utils"
)

func TestMsgSkippedEventTx_RawTxBytes(t *testing.T) {

	// A typical MsgSkippedEventTx sequence is always 2 or greater
	msgSkippedEventTxSequence := uint64(2)

	msgSkippedEventTx := testtypes.TestMsgSkippedEventTx
	msgSkippedEventTxAny, err := codectypes.NewAnyWithValue(msgSkippedEventTx)
	require.NoError(t, err)

	expectMsgSkippedEventTxBz, err := utils.ValidRawTxBytesFromAnyMsgs(
		[]*codectypes.Any{msgSkippedEventTxAny},
		msgSkippedEventTxSequence,
	)
	require.NoError(t, err)
	notExpectMsgSkippedEventTxBz1, err := utils.ValidRawTxBytesFromAnyMsgs(
		[]*codectypes.Any{msgSkippedEventTxAny},
		msgSkippedEventTxSequence+1,
	)
	require.NoError(t, err)
	notExpectMsgSkippedEventTxBz2, err := utils.ValidRawTxBytesFromAnyMsgs(
		[]*codectypes.Any{msgSkippedEventTxAny},
		msgSkippedEventTxSequence-1,
	)
	require.NoError(t, err)
	actualMsgSkippedEventTxRawBytes, err := msgSkippedEventTx.RawTxBytes(msgSkippedEventTxSequence)
	require.NoError(t, err)

	// The main point here is to ensure that the sequence is actually adhered to
	require.Equal(t, expectMsgSkippedEventTxBz, actualMsgSkippedEventTxRawBytes)
	require.NotEqual(t, notExpectMsgSkippedEventTxBz1, actualMsgSkippedEventTxRawBytes)
	require.NotEqual(t, notExpectMsgSkippedEventTxBz2, actualMsgSkippedEventTxRawBytes)
}
