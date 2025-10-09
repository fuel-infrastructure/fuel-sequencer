package sequencer

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"

	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/types"
	blobtypes "github.com/fuel-infrastructure/fuel-sequencer/x/blob/types"
)

// SubmitBlobMetadataTx posts blob data directly to the sequencer using x/blob module
func (c *Client) SubmitBlobMetadataTx(
	ctx context.Context, blobs []*types.TrackedBlob, txSequence uint64, blobNonce int) (*sdk.TxResponse, error) {
	// Create MsgsBlobMetadataTxs
	msgsPostBlobMetadata := make([]sdk.Msg, len(blobs))

	for i, blob := range blobs {
		msgsPostBlobMetadata[i] = &blobtypes.MsgBlobMetadataTx{
			Sender: c.Sender.Address,
			Hash:   blob.Receipt.Key.String(),
			Size_:  uint64(len(blob.StoredBlob.Data)),
			Topic:  c.topic,
			Nonce:  uint64(blobNonce + i),
		}
	}

	txFactory, err := c.setupTxFactory(msgsPostBlobMetadata, txSequence)
	if err != nil {
		return nil, fmt.Errorf("failed to setup transaction factory: %w", err)
	}

	// Create output buffer to capture transaction response
	outputBuffer := &bytes.Buffer{}
	clientCtxWithOutput := c.clientCtx.WithOutput(outputBuffer)

	// Broadcast transaction
	err = tx.GenerateOrBroadcastTxWithFactory(clientCtxWithOutput, txFactory, msgsPostBlobMetadata...)
	if err != nil {
		return nil, fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	// Verify transaction
	// Wait for response to be written to buffer
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	responseReceived := false
	for !responseReceived {
		select {
		case <-timeout:
			return nil, fmt.Errorf("timeout waiting for transaction response")
		case <-ticker.C:
			if outputBuffer.Len() > 0 {
				responseReceived = true
			}
		}
	}

	// Parse the transaction submission response (only has tx hash)
	var submissionResponse sdk.TxResponse
	err = c.clientCtx.Codec.UnmarshalJSON(outputBuffer.Bytes(), &submissionResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction response: %w", err)
	}
	if submissionResponse.Code != 0 {
		return &submissionResponse, fmt.Errorf("transaction failed: %s", submissionResponse.RawLog)
	}

	return &submissionResponse, nil
}

func (c *Client) setupTxFactory(msgs []sdk.Msg, txSequence uint64) (tx.Factory, error) {
	// Get the from address from client context (this is the actual derived address)
	fromAddr := c.clientCtx.GetFromAddress()

	// Query account to get current sequence
	acc, err := c.clientCtx.AccountRetriever.GetAccount(c.clientCtx, fromAddr)
	if err != nil {
		return tx.Factory{}, fmt.Errorf("failed to get account: %w", err)
	}

	// Create transaction factory
	txFactory := tx.Factory{}.
		WithAccountRetriever(c.clientCtx.AccountRetriever).
		WithChainID(c.chainID).
		WithTxConfig(c.clientCtx.TxConfig).
		WithKeybase(c.clientCtx.Keyring).
		WithAccountNumber(acc.GetAccountNumber()).
		WithSequence(txSequence).
		WithGasAdjustment(1.2).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)

	// Ensure account exists
	if err := txFactory.AccountRetriever().EnsureExists(c.clientCtx, fromAddr); err != nil {
		return txFactory, fmt.Errorf("failed to ensure account exists: %w", err)
	}

	_, gasEstimate, err := tx.CalculateGas(c.clientCtx, txFactory, msgs...)
	if err != nil {
		return txFactory, fmt.Errorf("failed to calculate gas: %w", err)
	}
	txFactory = txFactory.WithGas(gasEstimate)

	return txFactory, nil
}
