package abci

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	blobtypes "github.com/fuel-infrastructure/fuel-sequencer/x/blob/types"
)

// filterBlobTransactions filters blob transactions early in the process, before block space calculation
func (h *FuelSequencerProposalHandler) filterBlobTransactions(ctx sdk.Context, txs [][]byte) ([][]byte, error) {
	var validTxs [][]byte
	var blobTxs [][]byte

	for _, txBytes := range txs {
		tx, err := h.txVerifier.TxDecode(txBytes)
		if err != nil {
			// Skip transactions that can't be decoded
			ctx.Logger().Warn("skipping transaction that cannot be decoded", "error", err)
			continue
		}

		hasBlobMsg := false
		canInclude := true
		for _, msg := range tx.GetMsgs() {
			if blobMsg, ok := msg.(*blobtypes.MsgBlobMetadataTx); ok {
				hasBlobMsg = true
				hashStr := hex.EncodeToString(blobMsg.BlobHash)

				// Only check local blob pool - no external downloads
				if !h.blobKeeper.HasBlobData(ctx, blobMsg.BlobHash) {
					canInclude = false
					ctx.Logger().Debug("skipping blob transaction with unavailable blob in local pool", "hash", hashStr)
					break
				}
			}
		}

		if canInclude {
			if hasBlobMsg {
				blobTxs = append(blobTxs, txBytes)
			} else {
				validTxs = append(validTxs, txBytes)
			}
		}
	}

	// Return non-blob transactions first, then blob transactions
	return append(validTxs, blobTxs...), nil
}

// validateBlobTransactions validates that all blob transactions have available blobs and correct hashes
func (h *FuelSequencerProposalHandler) validateBlobTransactions(ctx sdk.Context, txs [][]byte) error {
	for _, txBytes := range txs {
		tx, err := h.txVerifier.TxDecode(txBytes)
		if err != nil {
			return fmt.Errorf("failed to decode transaction: %w", err)
		}

		for _, msg := range tx.GetMsgs() {
			if blobMsg, ok := msg.(*blobtypes.MsgBlobMetadataTx); ok {
				hashStr := hex.EncodeToString(blobMsg.BlobHash)

				// Only check local blob pool - no external downloads
				if !h.blobKeeper.HasBlobData(ctx, blobMsg.BlobHash) {
					return fmt.Errorf("blob not available in local pool: %s", hashStr)
				}

				// Get blob data from local pool and verify hash
				blobData, err := h.blobKeeper.GetBlobData(ctx, blobMsg.BlobHash)
				if err != nil {
					return fmt.Errorf("failed to get blob data from local pool: %s", hashStr)
				}

				// Verify hash matches
				computedHash := sha256.Sum256(blobData)
				if !bytes.Equal(computedHash[:], blobMsg.BlobHash) {
					return fmt.Errorf("blob hash mismatch: %s", hashStr)
				}
			}
		}
	}

	return nil
}

// countBlobTransactions counts the number of blob transactions in a slice
func (h *FuelSequencerProposalHandler) countBlobTransactions(ctx sdk.Context, txs [][]byte) (int, error) {
	count := 0

	for _, txBytes := range txs {
		tx, err := h.txVerifier.TxDecode(txBytes)
		if err != nil {
			continue
		}

		for _, msg := range tx.GetMsgs() {
			if _, ok := msg.(*blobtypes.MsgBlobMetadataTx); ok {
				count++
				break // Only count once per transaction
			}
		}
	}

	return count, nil
}
