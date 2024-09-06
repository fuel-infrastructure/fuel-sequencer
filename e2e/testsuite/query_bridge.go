package testsuite

import (
	"context"
	"fmt"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *E2ETestSuite) QueryBridgeParams(ctx context.Context) *bridgetypes.Params {
	queryClient := s.getGRPCClients().BridgeQueryClient
	res, err := queryClient.Params(ctx, &bridgetypes.QueryParamsRequest{})
	s.Require().NoError(err)

	return &res.Params
}

func (s *E2ETestSuite) QueryLastEthereumBlockSynced(ctx context.Context) uint64 {
	queryClient := s.getGRPCClients().BridgeQueryClient
	res, err := queryClient.LastEthereumBlockSynced(ctx, &bridgetypes.QueryGetLastEthereumBlockSyncedRequest{})
	s.Require().NoError(err)

	return res.Block
}

func (s *E2ETestSuite) PollForEthereumEventIndexOffset(
	ctx context.Context, deltaBlocks uint64, offset uint64,
) {
	h, err := s.Chain.FuelSequencerHeight(ctx)
	s.Require().NoError(err)

	s.T().Log(fmt.Sprintf("Polling for Ethereum event index offset %d", offset))

	doPoll := func(ctx context.Context, height uint64) (any, error) {
		resp, err := s.Chain.grpcClients.BridgeQueryClient.EthereumEventIndexOffset(ctx,
			&bridgetypes.QueryGetEthereumEventIndexOffsetRequest{},
		)
		if err != nil {
			return nil, err
		}
		if resp.Offset != offset {
			return nil, fmt.Errorf("offset (%d) does not match expected: (%d)", resp.Offset, offset)
		}
		return nil, nil
	}

	bp := BlockPoller[any]{CurrentHeight: s.Chain.FuelSequencerHeight, PollFunc: doPoll}
	_, err = bp.DoPoll(ctx, h, h+deltaBlocks)
	s.Require().NoError(err, fmt.Errorf("exact offset (%d) not found in expected number of blocks", offset))
}

func (s *E2ETestSuite) PollForLastEthereumBlockSynced(
	ctx context.Context, deltaBlocks uint64, block uint64,
) {
	h, err := s.Chain.FuelSequencerHeight(ctx)
	s.Require().NoError(err)

	s.T().Log(fmt.Sprintf("Polling for last Ethereum block synced %d", block))

	doPoll := func(ctx context.Context, height uint64) (any, error) {
		resp, err := s.Chain.grpcClients.BridgeQueryClient.LastEthereumBlockSynced(ctx,
			&bridgetypes.QueryGetLastEthereumBlockSyncedRequest{},
		)
		if err != nil {
			return nil, err
		}
		if resp.Block != block {
			return nil, fmt.Errorf("last Ethereum block synced (%d) does not match expected: (%d)", resp.Block, block)
		}
		return nil, nil
	}

	bp := BlockPoller[any]{CurrentHeight: s.Chain.FuelSequencerHeight, PollFunc: doPoll}
	_, err = bp.DoPoll(ctx, h, h+deltaBlocks)
	s.Require().NoError(err, fmt.Errorf("last Ethereum block synced %d not found in expected number of blocks", block))
}

func (s *E2ETestSuite) GetMsgIndexFromBlock(ctx context.Context, block int64) *bridgetypes.MsgIndex {
	s.Logger().Info(fmt.Sprintf("Getting MsgIndex from block %d", block))

	blockByHeight, err := s.GetBlockByHeight(ctx, block)
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(len(blockByHeight.Data.Txs), 1)

	txBz := blockByHeight.Data.Txs[0]
	sdkTx, err := authtx.DefaultTxDecoder(TestCdc)(txBz)
	s.Require().NoError(err)

	var msg bridgetypes.MsgIndex
	err = msg.FromSdkTx(sdkTx)
	s.Require().NoError(err)

	return &msg
}

func (s *E2ETestSuite) SearchForEventInBlockResults(
	ctx context.Context, eventType string, block int64,
) (event *abcitypes.Event, found bool) {
	s.Logger().Info(fmt.Sprintf("Looking for event %s at block %d", eventType, block))

	blockResults, err := s.GetBlockResultsByHeight(ctx, block)
	s.Require().NoError(err)

	for _, txResult := range blockResults.TxsResults {
		for _, event := range txResult.Events {
			if event.Type == eventType {
				return &event, true
			}
		}
	}
	return nil, false
}

func (s *E2ETestSuite) SearchForMsgInBlock(
	ctx context.Context, msgTypeUrl string, block int64,
) (msg sdk.Msg, txBz []byte, found bool) {
	s.Logger().Info(fmt.Sprintf("Looking for msg %s at block %d", msgTypeUrl, block))

	txDecoder := authtx.DefaultTxDecoder(TestCdc)

	blockByHeight, err := s.GetBlockByHeight(ctx, block)
	s.Require().NoError(err)

	for _, txBz := range blockByHeight.Data.Txs {
		sdkTx, err := txDecoder(txBz)
		s.Require().NoError(err)

		for _, msg := range sdkTx.GetMsgs() {
			if sdk.MsgTypeURL(msg) == msgTypeUrl {
				return msg, txBz, true
			}
		}
	}

	return nil, nil, false
}
