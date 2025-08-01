package abci

import (
	"context"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/fuel-infrastructure/fuel-sequencer/utils"
)

// fuelSequencerTxSelector is a custom implementation of the baseapp.TxSelector interface. It is largely based on
// baseapp.defaultTxSelector, with the only difference being the size calculation for transactions. We opted for this
// approach instead of waiting for the patch in the v0.50.0 Cosmos SDK series, as the Cosmos SDK team indicated that the
// change would only be included in a major release.
// Ref: https://github.com/cosmos/cosmos-sdk/pull/18551
type fuelSequencerTxSelector struct {
	totalTxBytes uint64
	totalTxGas   uint64
	selectedTxs  [][]byte
}

func NewFuelSequencerTxSelector() baseapp.TxSelector {
	return &fuelSequencerTxSelector{}
}

func (ts *fuelSequencerTxSelector) SelectedTxs(_ context.Context) [][]byte {
	txs := make([][]byte, len(ts.selectedTxs))
	copy(txs, ts.selectedTxs)
	return txs
}

func (ts *fuelSequencerTxSelector) Clear() {
	ts.totalTxBytes = 0
	ts.totalTxGas = 0
	ts.selectedTxs = nil
}

func (ts *fuelSequencerTxSelector) SelectTxForProposal(
	_ context.Context, maxTxBytes, maxBlockGas uint64, memTx sdk.Tx, txBz []byte,
) bool {
	txSize := utils.TxSize(txBz)

	var txGasLimit uint64
	if memTx != nil {
		if gasTx, ok := memTx.(baseapp.GasTx); ok {
			txGasLimit = gasTx.GetGas()
		}
	}

	// only add the transaction to the proposal if we have enough capacity
	if (txSize + ts.totalTxBytes) <= maxTxBytes {
		// If there is a max block gas limit, add the tx only if the limit has
		// not been met.
		if maxBlockGas > 0 {
			if (txGasLimit + ts.totalTxGas) <= maxBlockGas {
				ts.totalTxGas += txGasLimit
				ts.totalTxBytes += txSize
				ts.selectedTxs = append(ts.selectedTxs, txBz)
			}
		} else {
			ts.totalTxBytes += txSize
			ts.selectedTxs = append(ts.selectedTxs, txBz)
		}
	}

	// check if we've reached capacity; if so, we cannot select any more transactions
	return ts.totalTxBytes >= maxTxBytes || (maxBlockGas > 0 && (ts.totalTxGas >= maxBlockGas))
}
