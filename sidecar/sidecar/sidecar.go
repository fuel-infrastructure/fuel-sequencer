package sidecar

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v4"
	ethereumtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/ethwrappedclient"
	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/sequencerclient"
	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/store"
	"go.uber.org/zap"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
)

// Sidecar represents the sidecar service.
type Sidecar struct {
	logger *zap.Logger

	// Ethereum RPC client used for querying data from an Ethereum node.
	ethRpcClient *ethwrappedclient.EthRpcClient

	// Ethereum WS client used for subscribing to block headers from an Ethereum node.
	ethWsClient *ethwrappedclient.EthWsClient

	// SequencerClient is used for querying data from a Sequencer node.
	sequencerClient *sequencerclient.SequencerClient

	// eventStore stores all the necessary information needed to run the sidecar.
	eventStore *store.EventStore

	// stopped indicates if the main process of the sidecar has been stopped or not. By default, this is false.
	stopped atomic.Bool

	// fetchAndStoreLock makes fetching and storing of logs sequential to prevent duplicate queries if multiple blocks
	// are received rapidly, since the last synced block value from the previous fetch would not have been updated yet.
	fetchAndStoreLock sync.Mutex
}

// NewSidecar initializes a new Sidecar instance.
func NewSidecar(
	logger *zap.Logger,
	ethRpcClient *ethwrappedclient.EthRpcClient,
	ethWsClient *ethwrappedclient.EthWsClient,
	sequencerClient *sequencerclient.SequencerClient,
	eventStore *store.EventStore,
) *Sidecar {
	return &Sidecar{
		logger:          logger,
		ethRpcClient:    ethRpcClient,
		ethWsClient:     ethWsClient,
		sequencerClient: sequencerClient,
		eventStore:      eventStore,
	}
}

// Start begins the sidecar's operation of fetching Ethereum logs, with an initial connectivity check.
func (s *Sidecar) Start(ctx context.Context) error {
	defer s.ShutDown()

	s.logger.Info("starting log fetching")

	// TODO: Do check for rpc as well

	// Initial check to verify Ethereum client connectivity and log subscription capability.
	// Note: if a non-websocket URL is provided, this check will fail as well.
	sub, err := s.ethClient.SubscribeNewHead(context.Background(), make(chan *ethereumtypes.Header))
	if err != nil {
		s.logger.Error("failed initial Ethereum subscription check", zap.Error(err))
		return err
	}
	sub.Unsubscribe()

	return s.startFetchingLogs(ctx)
}

// getMaxSyncableBlock gets the last Ethereum height that the Sidecar can sync up to. In normal operation, this will be
// the height of the last finalized Ethereum block. However, if an end query block is available, the end query block
// will override the finalized Ethereum block.
func (s *Sidecar) getMaxSyncableBlock(ctx context.Context) (*big.Int, error) {

	// Get the height of the last finalized block.
	finalizedEthHeight, err := s.ethClient.FinalizedBlockNumber(ctx)
	if err != nil {
		s.logger.Error("could not get the height of the last finalized block from Ethereum node", zap.Error(err))
		return nil, err
	}

	// If the end query block exists and is less than the finalized Ethereum height, use it instead.
	endQueryBlock := s.eventStore.GetEndQueryBlock()
	if endQueryBlock != nil && endQueryBlock.Cmp(finalizedEthHeight) < 0 {
		s.logger.Warn(
			"using end query block instead of finalized Ethereum height",
			zap.Uint64("finalized_eth_height", finalizedEthHeight.Uint64()),
			zap.Uint64("end_query_block", endQueryBlock.Uint64()),
		)
		return endQueryBlock, nil
	}

	return finalizedEthHeight, nil
}

// startFetchingLogs fetches logs from Ethereum and processes them, with a catch-up process to sync historical logs.
func (s *Sidecar) startFetchingLogs(ctx context.Context) error {

	// Configure backoff for Ethereum logs subscription.
	backOff := backoff.NewExponentialBackOff()
	backOff.InitialInterval = time.Second
	backOff.MaxInterval = 5 * time.Second
	backOff.RandomizationFactor = 0 // no funny business
	backOff.MaxElapsedTime = 0      // never stop retrying
	backOff.Reset()

	// Set up the operation to fetch logs from Ethereum.
	// Note: returning an error inside means we should retry.
	//
	// Two special cases:
	// - If a retry was requested with a nil error, a dummy error is returned to guarantee the retry.
	// - If an error is returned with no retry request, PermanentError is used to guarantee no retry.
	operation := backoff.Operation(func() error {
		if err := s.catchUpWithEthereumLogs(ctx, backOff); err != nil {
			return err
		}

		err, retry := s.subscribeToNewEthereumLogs(ctx, backOff)
		if retry {
			// Retry
			if err != nil {
				return err
			} else {
				return fmt.Errorf("retrying with no error")
			}
		} else {
			// No retry
			if err != nil {
				return &backoff.PermanentError{Err: err}
			} else {
				return nil
			}
		}
	})

	// Set up a notify function to log the retry.
	notify := backoff.Notify(func(err error, duration time.Duration) {
		s.logger.Debug("retrying after backoff delay", zap.Duration("delay", duration))
	})

	// Run the operation with the configured backoff.
	err := backoff.RetryNotify(operation, backOff, notify)
	if err != nil {
		s.logger.Error("error from backoff retry", zap.Error(err))
	}

	return nil
}

// catchUpWithEthereumLogs syncs logs from the last synced block up to the last finalized Ethereum block.
func (s *Sidecar) catchUpWithEthereumLogs(
	ctx context.Context, backOff *backoff.ExponentialBackOff,
) error {

	// Get the max syncable block (considers finalized Ethereum height and the end query block)
	maxSyncableBlock, err := s.getMaxSyncableBlock(ctx)
	if err != nil {
		return err
	}

	// If we're behind, fetch the logs up till the max syncable block.
	// Note: in the meantime more Ethereum blocks might be finalized, but, these can be detected and fetched in the
	// main data fetching loop.
	lastSyncedBlock := s.eventStore.GetLastSyncedBlock()
	if maxSyncableBlock.Cmp(lastSyncedBlock) > 0 {
		s.logger.Info("catching up with ethereum",
			zap.Uint64("last_synced_block", lastSyncedBlock.Uint64()),
			zap.Uint64("max_syncable_block", maxSyncableBlock.Uint64()),
			zap.Uint64("max_query_range", s.eventStore.GetMaxQueryRange().Uint64()),
		)

		return s.fetchAndStoreLogsUptoBlock(ctx, maxSyncableBlock)
	} else {
		s.logger.Debug("already in sync with ethereum",
			zap.Uint64("last_synced_block", lastSyncedBlock.Uint64()),
			zap.Uint64("max_syncable_block", maxSyncableBlock.Uint64()),
		)
	}

	// Reset the backoff just in case we've used it. Without this, the backoff interval does not get reset.
	backOff.Reset()

	return nil
}

// subscribeToNewEthereumLogs syncs logs from newly finalized Ethereum blocks.
func (s *Sidecar) subscribeToNewEthereumLogs(
	ctx context.Context, backOff *backoff.ExponentialBackOff,
) (err error, retry bool) {
	s.logger.Info("subscribing to new ethereum block headers",
		zap.Uint64("last_synced_block", s.eventStore.GetLastSyncedBlock().Uint64()),
	)

	// Subscribe to new Ethereum block headers.
	ch := make(chan *ethereumtypes.Header)
	sub, err := s.ethClient.SubscribeNewHead(ctx, ch)
	if err != nil {
		s.logger.Error("error when subscribing to logs", zap.Error(err))
		return err, true // retry
	}

	for {
		select {
		case <-ctx.Done():
			s.logger.Warn("sidecar stopped via context", zap.Error(ctx.Err()))
			sub.Unsubscribe()
			return ctx.Err(), false // no retry
		case err := <-sub.Err():
			s.logger.Error("error from logs subscription", zap.Error(err))
			sub.Unsubscribe()
			return err, true // retry
		case header := <-ch:

			// If the sidecar has been stopped, exit.
			if s.IsStopped() {
				// TODO: Do Unsubscribe jic?
				return fmt.Errorf("received new header but sidecar is stopped"), false // no retry
			}

			// Get the max syncable block (considers finalized Ethereum height and the end query block)
			maxSyncableBlock, err := s.getMaxSyncableBlock(ctx)
			if err != nil {
				return err, true
			}

			// Get the last Ethereum block synced by the Sidecar
			lastSyncedBlock := s.eventStore.GetLastSyncedBlock()

			s.logger.Info("detected new block header",
				zap.Uint64("last_synced_block", lastSyncedBlock.Uint64()),
				zap.Uint64("max_syncable_block", maxSyncableBlock.Uint64()),
				zap.Uint64("detected_eth_height", header.Number.Uint64()),
				zap.Uint64("max_query_range", s.eventStore.GetMaxQueryRange().Uint64()),
			)

			// If new blocks are syncable, process all logs between the last synced blocked and max syncable block.
			if maxSyncableBlock.Cmp(lastSyncedBlock) > 0 {
				err := s.fetchAndStoreLogsUptoBlock(ctx, maxSyncableBlock)
				if err != nil {
					return err, true // retry
				}
			}

			// Fetch the last synced Ethereum block before querying for new logs
			lastSyncedBlockBySequencer, err := s.sequencerClient.FetchLastEthereumBlockSynced(ctx)
			if err != nil {
				// Log the error if the last synced Ethereum block is not obtained.
				// Note; We should still attempt to process Ethereum blocks. Reason being is that if the processing
				// is skipped the Sequencer will not be able to produce the first block and the sidecar would not be
				// able to query the Sequencer, causing a deadlock.
				s.logger.Error("failed to obtain LastEthereumBlockSynced from Sequencer", zap.Error(err))
			} else {
				s.logger.Debug(
					"queried LastEthereumBlockSynced from Sequencer",
					zap.String("block", lastSyncedBlockBySequencer.String()),
				)
			}

			// Prune any old events that are no longer necessary to keep.
			s.eventStore.CalibrateBlocksAndPruneLogs(s.logger, lastSyncedBlockBySequencer)

			// Reset the backoff just in case we've used it. Without this, the backoff interval does not get reset.
			backOff.Reset()
		}
	}
}

// fetchAndProcessLogs fetches and processes logs based on next block to query, target block, and query max range.
func (s *Sidecar) fetchAndStoreLogsUptoBlock(ctx context.Context, toBlock *big.Int) error {
	s.fetchAndStoreLock.Lock()
	defer s.fetchAndStoreLock.Unlock()

	for {
		// Fetch logs (note: this is rate-limited under the hood)
		eventsMap, newLastSyncedBlock, err := s.ethClient.FetchAndProcessLogs(
			ctx, s.eventStore.GetNextQueryBlock(), toBlock, s.eventStore.GetMaxQueryRange(),
		)
		if err != nil {
			s.logger.Error("error fetching logs", zap.Error(err))
			return err
		}

		// Store the newly fetched events and update the last synced block.
		s.eventStore.AddEvents(eventsMap)
		s.eventStore.SetLastSyncedBlock(newLastSyncedBlock)

		// If we've reached the requested block, we can return.
		if newLastSyncedBlock.Cmp(toBlock) == 0 {
			return nil
		}
	}
}

// QueryBlockEvents queries the `blocksMap` for events associated with a specific block number.
func (s *Sidecar) QueryBlockEvents(blockNumber *big.Int) ([]sidecartypes.Event, error) {
	s.logger.Debug("processing block events query", zap.String("block", blockNumber.String()))

	// Validate the blockNumber
	if blockNumber.Sign() < 0 {
		return nil, errors.New("block number cannot be negative")
	}

	events, exists := s.eventStore.GetStoredEvents(blockNumber)
	if !exists {

		startQueryBlock := s.eventStore.GetStartQueryBlock()
		nextQueryBlock := s.eventStore.GetNextQueryBlock()

		// If the queried block is in the range of blocks saved in state but no events were found, return an empty list.
		// Example: if start block is 90 and next query block is 101, if there is no blocksMap entry for a query between
		//          90 and 100, this means that the queried block had no events.
		if (startQueryBlock != nil && nextQueryBlock != nil) &&
			(blockNumber.Cmp(startQueryBlock) >= 0 && blockNumber.Cmp(nextQueryBlock) < 0) {
			return []sidecartypes.Event{}, nil
		}

		// If the queried block is before the range of blocks saved in state, the state has been pruned.
		// Example: if start block is 90 then we know that we do not have the data for 89 and before.
		if startQueryBlock != nil && blockNumber.Cmp(startQueryBlock) < 0 {
			return nil, fmt.Errorf("block %s was pruned or never fetched", blockNumber.String())
		}

		// Otherwise this block was not yet processed
		return nil, fmt.Errorf("block %s not yet processed", blockNumber.String())
	}

	return events, nil
}

// IsStopped returns true if the sidecar has been stopped.
func (s *Sidecar) IsStopped() bool {
	return s.stopped.Load()
}

// ShutDown signals the sidecar to stop.
func (s *Sidecar) ShutDown() {
	s.logger.Warn("shutting down sidecar")
	s.stopped.Store(true)

	// RPC connection needs to be closed since we have already dialed.
	s.logger.Warn("closing RPC connection with Ethereum node")
	s.ethRpcClient.Close()

	// WS connection needs to be closed since we have already dialed.
	s.logger.Warn("closing WS connection with Ethereum node")
	s.ethWsClient.Close()
}
