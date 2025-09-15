package sequencer

import (
	"context"
	"fmt"
	"strings"
	"time"

	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/fuel-infrastructure/fuel-sequencer/testutil"
)

const (
	BlockTime      = 6 * time.Second
	BlockRetention = 2 * BlockTime // 2 blocks
)

// waitForTransactionConfirmation waits for a transaction to be included in a block and returns the full response
func (c *Client) WaitForMetadataTxFinalisation(ctx context.Context, txHash string) (*coretypes.ResultTx, error) {
	timer := time.NewTimer(BlockRetention)
	defer timer.Stop()

	pendingErrorString := fmt.Sprintf("tx (%s) not found", txHash)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timer.C:
			return nil, fmt.Errorf("timeout waiting for transaction confirmation")
		default:
			// Query the transaction to get full details
			txResponse, err := c.queryTransaction(ctx, txHash)
			if err != nil {
				if !strings.Contains(err.Error(), pendingErrorString) {
					return nil, err
				}
			} else if txResponse.Height > 0 {
				// Transaction has been included in a block
				return txResponse, nil
			}
			// Continue waiting if transaction not found yet or not included
			time.Sleep(1 * time.Second)
		}
	}
}

// queryTransaction queries a transaction by hash and returns the full response
func (c *Client) queryTransaction(ctx context.Context, txHash string) (*coretypes.ResultTx, error) {
	// Use the RPC client to query the transaction
	rpcClient := c.clientCtx.Client

	// Query transaction by hash
	txResult, err := rpcClient.Tx(ctx, testutil.MustHexDecodeString(txHash), false)
	if err != nil {
		return nil, err
	}
	if txResult.TxResult.Code != 0 {
		return nil, fmt.Errorf("transaction failed with code %d", txResult.TxResult.Code)
	}

	return txResult, nil
}
