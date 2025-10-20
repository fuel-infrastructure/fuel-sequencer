package utils

import cmttypes "github.com/cometbft/cometbft/types"

// TxSize returns the total amount of bytes that a transaction occupies. We are using cmttypes.ComputeProtoSizeForTxs
// as opposed to len(txBz) to accurately measure the size of the transaction when serialized by CometBFT.
// Ref: https://github.com/cosmos/cosmos-sdk/pull/18551
func TxSize(txBz []byte) uint64 {
	return uint64(cmttypes.ComputeProtoSizeForTxs([]cmttypes.Tx{txBz})) //nolint:gosec // Safe conversion, size is small
}

// TxsSize returns the total amount of bytes that a set of transactions occupy. This makes use of TxSize to accurately
// measure the size of the transactions when serialized by CometBFT.
func TxsSize(txsBz [][]byte) (size uint64) {
	for _, txBz := range txsBz {
		size += TxSize(txBz)
	}

	return size
}
