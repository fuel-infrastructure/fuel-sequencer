package types_test

import (
	"strings"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/utils"
	"github.com/stretchr/testify/require"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func TestMsgSkippedEventTx_RawTxBytes(t *testing.T) {

	// A typical MsgSkippedEventTx sequence is always 2 or greater
	msgSkippedEventTxSequence := uint64(2)

	msgSkippedEventTx := testtypes.TestMsgSkippedEventTx
	msgSkippedEventTxAny, err := codectypes.NewAnyWithValue(msgSkippedEventTx)
	require.NoError(t, err)

	expectMsgSkippedEventTxBz, err := utils.ValidRawTxBytesFromAnyMsgs([]*codectypes.Any{msgSkippedEventTxAny}, msgSkippedEventTxSequence)
	require.NoError(t, err)
	notExpectMsgSkippedEventTxBz1, err := utils.ValidRawTxBytesFromAnyMsgs([]*codectypes.Any{msgSkippedEventTxAny}, msgSkippedEventTxSequence+1)
	require.NoError(t, err)
	notExpectMsgSkippedEventTxBz2, err := utils.ValidRawTxBytesFromAnyMsgs([]*codectypes.Any{msgSkippedEventTxAny}, msgSkippedEventTxSequence-1)
	require.NoError(t, err)
	actualMsgSkippedEventTxRawBytes, err := msgSkippedEventTx.RawTxBytes(msgSkippedEventTxSequence)
	require.NoError(t, err)

	// The main point here is to ensure that the sequence is actually adhered to
	require.Equal(t, expectMsgSkippedEventTxBz, actualMsgSkippedEventTxRawBytes)
	require.NotEqual(t, notExpectMsgSkippedEventTxBz1, actualMsgSkippedEventTxRawBytes)
	require.NotEqual(t, notExpectMsgSkippedEventTxBz2, actualMsgSkippedEventTxRawBytes)
}

func TestMsgSkippedEventTx_StringLength(t *testing.T) {
	tests := []struct {
		name           string
		reason         string
		expectedLength int
		shouldTrim     bool
	}{
		{
			name:           "reason within limit",
			reason:         "normal reason",
			expectedLength: 13, // "normal reason" is 13 characters
			shouldTrim:     false,
		},
		{
			name:           "reason at limit",
			reason:         strings.Repeat("a", types.MaxReasonLength),
			expectedLength: types.MaxReasonLength,
			shouldTrim:     false,
		},
		{
			name:           "reason exceeds limit",
			reason:         strings.Repeat("a", types.MaxReasonLength+1),
			expectedLength: types.MaxReasonLength,
			shouldTrim:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, trimmed := types.NewMsgSkippedEventTx(
				"authority",
				tt.reason,
				1,       // ethBlockNumber
				2,       // ethLogIndex
				3,       // ethTxIndex
				"0x123", // ethTxHash
			)

			require.Equal(t, tt.expectedLength, len(msg.ReasonForSkip))
			require.Equal(t, tt.shouldTrim, trimmed)
			if tt.shouldTrim {
				require.True(t, strings.HasSuffix(msg.ReasonForSkip, "..."))
			} else {
				require.Equal(t, tt.reason, msg.ReasonForSkip)
			}
		})
	}
}
