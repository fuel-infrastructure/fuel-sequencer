package sequencer

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/types"
	sequencingtypes "github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"
)

// generateUniqueTopicID creates a unique topic ID by appending timestamp to avoid conflicts
func generateUniqueTopicID(topic string) []byte {
	// Append current timestamp to make topic unique for each run
	uniqueTopic := fmt.Sprintf("%s-%d", topic, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(uniqueTopic))
	return hash[:]
}

// SubmitPostBlob posts blob data directly to the sequencer using MsgPostBlob
func (c *Client) SubmitPostBlob(
	ctx context.Context, blobs []*types.TrackedBlob, txSequence uint64, blobNonce int) (*sdk.TxResponse, error) {
	// Generate a unique topic ID for this run to avoid conflicts
	topicID := generateUniqueTopicID(c.topic)

	// Submit blobs one at a time with sequential ordering starting from 0
	var lastResponse *sdk.TxResponse
	for i, blob := range blobs {
		// Since we're using a unique topic ID, we can start with order 0 and increment
		order := math.NewInt(int64(i))

		// Create MsgPostBlob message for this single blob
		msgPostBlob := &sequencingtypes.MsgPostBlob{
			From:  c.Sender.Address,
			Topic: topicID,
			Order: order,
			Data:  blob.StoredBlob.Data,
		}

		// Submit this single blob
		response, err := c.submitSingleBlob(ctx, msgPostBlob, txSequence+uint64(i))
		if err != nil {
			return nil, fmt.Errorf("failed to submit blob %d: %w", i, err)
		}

		lastResponse = response
	}

	return lastResponse, nil
}

// submitSingleBlob submits a single MsgPostBlob message
func (c *Client) submitSingleBlob(ctx context.Context, msg *sequencingtypes.MsgPostBlob, txSequence uint64) (*sdk.TxResponse, error) {
	// Submit single blob to sequencer
	txFactory, err := c.setupTxFactory([]sdk.Msg{msg}, txSequence)
	if err != nil {
		return nil, fmt.Errorf("failed to setup transaction factory: %w", err)
	}

	// Create output buffer to capture transaction response
	outputBuffer := &bytes.Buffer{}
	clientCtxWithOutput := c.clientCtx.WithOutput(outputBuffer)

	// Broadcast transaction
	err = tx.GenerateOrBroadcastTxWithFactory(clientCtxWithOutput, txFactory, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	// Wait for response to be written to buffer
	timeout := time.After(15 * time.Second) // Increased timeout
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	responseReceived := false
	for !responseReceived {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context canceled while waiting for transaction response: %w", ctx.Err())
		case <-timeout:
			return nil, fmt.Errorf("timeout waiting for transaction response after 15 seconds")
		case <-ticker.C:
			if outputBuffer.Len() > 0 {
				responseReceived = true
			}
		}
	}

	// Parse the transaction submission response
	var submissionResponse sdk.TxResponse
	err = c.clientCtx.Codec.UnmarshalJSON(outputBuffer.Bytes(), &submissionResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction response: %w", err)
	}

	if submissionResponse.Code != 0 {
		return &submissionResponse, fmt.Errorf("transaction failed with code %d: %s", submissionResponse.Code, submissionResponse.RawLog)
	}

	return &submissionResponse, nil
}
