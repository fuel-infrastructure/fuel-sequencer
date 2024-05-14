package sidecar

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff"
	ethereumtypes "github.com/ethereum/go-ethereum/core/types"
	ethclient "github.com/fuel-infrastructure/fuel-sequencer/sidecar/ethwrappedclient"
	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/sequencerclient"
	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/store"
	"go.uber.org/zap"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
)

// Sidecar represents the sidecar service.
type Sidecar struct {
	logger *zap.Logger

	// Ethereum client used for querying data from an Ethereum node.
	ethClient *ethclient.EthWrappedClient

	// SequencerClient is used for querying data from a Sequencer node.
	sequencerClient *sequencerclient.SequencerClient

	// eventStore stores all the necessary information needed to run the sidecar.
	eventStore *store.EventStore

	// stopped indicates if the main process of the sidecar has been stopped or not. By default, this is false.
	stopped atomic.Bool

	// development is a boolean variable which indicates whether the sidecar should be run in development mode. If true
	// the sidecar will possess a development logger and will bypass some logic which is not available in development
	// mode. Ex. net_peerCount method is not available on anvil nodes, so logic around it needs to be bypassed in
	// development mode.
	development bool

	// fetchAndStoreLock makes fetching and storing of logs sequential to prevent duplicate queries if multiple blocks
	// are received rapidly, since the last synced block value from the previous fetch would not have been updated yet.
	fetchAndStoreLock sync.Mutex
}

// NewSidecar initializes a new Sidecar instance.
func NewSidecar(
	logger *zap.Logger,
	ethClient *ethclient.EthWrappedClient,
	sequencerClient *sequencerclient.SequencerClient,
	eventStore *store.EventStore,
	development bool,
) *Sidecar {
	return &Sidecar{
		logger:          logger,
		ethClient:       ethClient,
		sequencerClient: sequencerClient,
		eventStore:      eventStore,
		development:     development,
	}
}

// Start begins the sidecar's operation of fetching Ethereum logs, with an initial connectivity check.
func (s *Sidecar) Start(ctx context.Context) error {
	defer s.ShutDown()

	s.logger.Info("starting log fetching")

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

// startFetchingLogs fetches logs from Ethereum and processes them, with a catch-up process to sync historical logs.
func (s *Sidecar) startFetchingLogs(ctx context.Context) error {

	// Configure backoff for Ethereum logs subscription.
	backOff := backoff.NewExponentialBackOff()
	backOff.InitialInterval = time.Second
	backOff.MaxInterval = 5 * time.Second
	backOff.RandomizationFactor = 0 // no funny business
	backOff.Reset()

	// Set up the operation to fetch logs from Ethereum.
	// Note: returning an error inside means we should retry.
	operation := backoff.Operation(func() error {
		if err := s.catchUpWithEthereumLogs(ctx); err != nil {
			return err
		}
		if err, retry := s.subscribeToNewEthereumLogs(ctx); retry {
			return err
		}
		return nil
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
func (s *Sidecar) catchUpWithEthereumLogs(ctx context.Context) error {

	// Get the height of the last finalized block to determine whether we need to sync up.
	finalizedEthHeightUint64, err := s.ethClient.FinalizedBlockNumber(ctx)
	if err != nil {
		s.logger.Error("could not get the height of the last finalized block from Ethereum node", zap.Error(err))
		return err
	}
	finalizedEthHeight := new(big.Int).SetUint64(finalizedEthHeightUint64)

	// If we're behind, fetch the logs up till the last finalized Ethereum block.
	// Note: in the meantime more Ethereum blocks might be finalized, but, these can be detected and fetched in the
	// main data fetching loop.
	lastSyncedBlock := s.eventStore.GetLastSyncedBlock()
	if finalizedEthHeight.Cmp(lastSyncedBlock) > 0 {
		s.logger.Info("catching up with ethereum",
			zap.Uint64("last_synced_block", lastSyncedBlock.Uint64()),
			zap.Uint64("finalized_eth_height", finalizedEthHeightUint64),
			zap.Uint64("max_query_range", s.eventStore.GetMaxQueryRange().Uint64()),
		)

		return s.fetchAndStoreLogsUptoBlock(ctx, finalizedEthHeight)
	} else {
		s.logger.Debug("already in sync with ethereum",
			zap.Uint64("last_synced_block", lastSyncedBlock.Uint64()),
			zap.Uint64("finalized_eth_height", finalizedEthHeightUint64),
		)
	}

	return nil
}

// subscribeToNewEthereumLogs syncs logs from newly finalized Ethereum blocks.
func (s *Sidecar) subscribeToNewEthereumLogs(ctx context.Context) (err error, retry bool) {
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
				return fmt.Errorf("received new header but sidecar is stopped"), false // no retry
			}

			// Get the height of the last finalized block to determine whether a new finalized block needs processing
			finalizedEthHeightUint64, err := s.ethClient.FinalizedBlockNumber(ctx)
			if err != nil {
				s.logger.Error(
					"could not get the height of the last finalized block from Ethereum node", zap.Error(err),
				)
				return err, true
			}
			finalizedEthHeight := new(big.Int).SetUint64(finalizedEthHeightUint64)

			// Get the last Ethereum block synced by the Sidecar
			lastSyncedBlock := s.eventStore.GetLastSyncedBlock()

			s.logger.Info("detected new block header",
				zap.Uint64("last_synced_block", lastSyncedBlock.Uint64()),
				zap.Uint64("finalized_eth_height", finalizedEthHeightUint64),
				zap.Uint64("detected_eth_height", header.Number.Uint64()),
				zap.Uint64("max_query_range", s.eventStore.GetMaxQueryRange().Uint64()),
			)

			// If new blocks have been finalized, process all logs between the last synced blocked and the last
			// finalized block
			if finalizedEthHeight.Cmp(lastSyncedBlock) > 0 {
				err := s.fetchAndStoreLogsUptoBlock(ctx, finalizedEthHeight)
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
	s.stopped.Store(true)
}
