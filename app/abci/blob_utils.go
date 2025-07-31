package abci

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/blob-storage/pkg/store"
	blobtypes "github.com/fuel-infrastructure/fuel-sequencer/x/blob/types"
)

// filterBlobTransactions filters blob transactions early in the process, before block space calculation
func (h *FuelSequencerProposalHandler) filterBlobTransactions(ctx sdk.Context, txs [][]byte) ([][]byte, error) {
	var otherTxs [][]byte
	var blobTxs [][]byte

	for _, txBytes := range txs {
		tx, err := h.txVerifier.TxDecode(txBytes)
		if err != nil {
			// Skip transactions that can't be decoded
			ctx.Logger().Warn("transaction cannot be decoded - assuming non blob metadata tx", "error", err)
			otherTxs = append(otherTxs, txBytes)
			continue
		}

		hasBlobMsg := false
		canInclude := true
		for _, msg := range tx.GetMsgs() {
			if blobMsg, ok := msg.(*blobtypes.MsgBlobMetadataTx); ok {
				hasBlobMsg = true
				hash, err := store.ParseKey(blobMsg.Hash)
				if err != nil {
					return nil, fmt.Errorf("failed to parse blob hash: %w", err)
				}

				// Only check local blob pool - no external downloads
				if !h.blobKeeper.Has(hash) {
					canInclude = false
					ctx.Logger().Debug("skipping blob transaction with unavailable blob in local pool", "hash", hash)
					break
				}
			}
		}

		if canInclude {
			if hasBlobMsg {
				blobTxs = append(blobTxs, txBytes)
			} else {
				otherTxs = append(otherTxs, txBytes)
			}
		}
	}

	// Return non-blob transactions first, then blob transactions
	return append(otherTxs, blobTxs...), nil
}

// validateBlobTransactions validates that all blob transactions have available blobs and correct hashes
func (h *FuelSequencerProposalHandler) validateBlobTransactions(txs [][]byte) error {
	for _, txBytes := range txs {
		tx, err := h.txVerifier.TxDecode(txBytes)
		if err != nil {
			return fmt.Errorf("failed to decode transaction: %w", err)
		}

		for _, msg := range tx.GetMsgs() {
			if blobMsg, ok := msg.(*blobtypes.MsgBlobMetadataTx); ok {
				metadataKey, err := store.ParseKey(blobMsg.Hash)
				if err != nil {
					return fmt.Errorf("failed to parse blob hash: %w", err)
				}

				// Only check local blob pool - no external downloads
				if !h.blobKeeper.Has(metadataKey) {
					return fmt.Errorf("blob not available in local pool: %s", metadataKey)
				}

				// Get blob data from local pool and verify hash
				blob, err := h.blobKeeper.Get(metadataKey)
				if err != nil {
					return fmt.Errorf("failed to get blob data from local pool: %s", metadataKey)
				}

				// Verify hash matches
				recomputedKey := store.NewKey(blob.Data)
				if recomputedKey.String() != metadataKey.String() {
					return fmt.Errorf("blob hash mismatch: %s", metadataKey)
				}
			}
		}
	}

	return nil
}
