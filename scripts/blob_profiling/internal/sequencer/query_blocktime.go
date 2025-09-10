package sequencer

import (
	"context"
	"time"
)

func (c *Client) QueryBlockTime(ctx context.Context, height int64) (time.Time, error) {
	rpcClient := c.clientCtx.Client

	blockResult, err := rpcClient.Block(ctx, &height)
	if err != nil {
		return time.Time{}, err
	}

	return blockResult.Block.Header.Time, nil
}
