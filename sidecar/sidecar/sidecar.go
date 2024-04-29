package sidecar

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	ethereumtypes "github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"

	ethclient "github.com/fuel-infrastructure/fuel-sequencer/sidecar/ethwrappedclient"
	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/sequencerclient"
	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/store"

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

	// updateInterval is the wait between each extraction of logs when doing successive extractions in a short time.
	// TODO: improve use of updateInterval
	updateInterval time.Duration

	// stopped indicates if the main process of the sidecar has been stopped or not.
	stopped atomic.Bool

	// development is a boolean variable which indicates whether the sidecar should be run in development mode. If true
	// the sidecar will possess a development logger and will bypass some logic which is not available in development
	// mode. Ex. net_peerCount method is not available on anvil nodes, so logic around it needs to be bypassed in
	// development mode.
	development bool

	// acceptableDelay is the number of blocks that the Sidecar can be out-of-sync with Ethereum. When this occurs the
	// sidecar returns a special error message if a block which has not been processed yet is queried.
	acceptableDelay uint64

	// fetchAndStoreLock makes fetching and storing of logs sequential to prevent duplicate queries if multiple blocks
	// are received rapidly, since the last synced block value from the previous fetch would not have been updated yet.
	fetchAndStoreLock sync.Mutex

	// TODO: add consideration of block finality
}

// NewSidecar initializes a new Sidecar instance.
func NewSidecar(
	logger *zap.Logger,
	ethClient *ethclient.EthWrappedClient,
	sequencerClient *sequencerclient.SequencerClient,
	eventStore *store.EventStore,
	development bool,
	acceptableDelay uint64,
) *Sidecar {
	return &Sidecar{
		logger:          logger,
		ethClient:       ethClient,
		sequencerClient: sequencerClient,
		eventStore:      eventStore,
		updateInterval:  10 * time.Second,
		development:     development,
		acceptableDelay: acceptableDelay,
	}
}

// StartFetching begins the sidecar's operation.
func (s *Sidecar) StartFetching(ctx context.Context) error {
	s.logger.Info("starting data fetching")

	// Initial check to verify Ethereum client connectivity and log fetching capability.
	_, err := s.ethClient.FilterLogs(ctx, s.eventStore.GetStartQueryBlock(), s.eventStore.GetStartQueryBlock())
	if err != nil {
		s.logger.Error("failed to fetch logs for initial check", zap.Error(err))
		return err
	}

	go s.startFetching(ctx)

	return nil
}

// startFetching continuously fetches logs from the Ethereum blockchain and processes them.
func (s *Sidecar) startFetching(ctx context.Context) {
	s.stopped.Store(false)
	defer s.stopped.Store(true)

	// Get the current Ethereum height to determine if we already need to sync up.
	ethHeightUint64, err := s.ethClient.BlockNumber(ctx)
	if err != nil {
		s.logger.Error("could not get latest height from Ethereum node", zap.Error(err))
		return
	}
	ethHeight := new(big.Int).SetUint64(ethHeightUint64)

	// If we're behind, fetch the logs up till the current Ethereum height.
	// Note: in the meantime, more Ethereum blocks will be created, but
	// these can be detected and fetched in the main data fetching loop.
	lastSyncedBlock := s.eventStore.GetLastSyncedBlock()
	if ethHeight.Cmp(lastSyncedBlock) > 0 {
		s.logger.Info("catching up with ethereum",
			zap.Uint64("last_synced_block", lastSyncedBlock.Uint64()),
			zap.Uint64("eth_height", ethHeightUint64),
		)

		s.fetchAndStoreLogsUptoBlock(ctx, ethHeight)
	}

	s.logger.Info("subscribing to new ethereum block headers")

	// Subscribe to new Ethereum block headers.
	ch := make(chan *ethereumtypes.Header)
	sub, err := s.ethClient.SubscribeNewHead(ctx, ch)
	if err != nil {
		s.logger.Error("error when subscribing to logs", zap.Error(err))
		return
	}
	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			s.ShutDown()
			s.logger.Info("sidecar stopped via context")
		case err := <-sub.Err():
			s.logger.Error("error from logs subscription", zap.Error(err))
			// TODO: are there cases where we want to recreate the subscription here?
		case header := <-ch:

			// If the sidecar has been stopped, exit.
			if s.IsStopped() {
				return
			}

			// Fetch and store events
			s.logger.Info("detected new block header", zap.Uint64("block", header.Number.Uint64()))
			s.fetchAndStoreLogsUptoBlock(ctx, header.Number)

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
			s.eventStore.PruneLogs(s.logger, lastSyncedBlockBySequencer)
		}
	}
}

// fetchAndProcessLogs fetches and processes logs based on the next query, latest block, and the query max range.
func (s *Sidecar) fetchAndStoreLogsUptoBlock(ctx context.Context, toBlock *big.Int) {
	s.fetchAndStoreLock.Lock()
	defer s.fetchAndStoreLock.Unlock()

	for {
		eventsMap, newLastSyncedBlock, err := s.ethClient.FetchAndProcessLogs(
			ctx, s.eventStore.GetNextQueryBlock(), toBlock, s.eventStore.GetMaxQueryRange(),
		)
		if err != nil {
			s.logger.Error("error fetching logs", zap.Error(err))
			return
		}

		// Store the newly fetched events and update the last synced block.
		s.eventStore.AddEvents(eventsMap)
		s.eventStore.SetLastSyncedBlock(newLastSyncedBlock)

		if newLastSyncedBlock.Cmp(toBlock) == 0 {
			return
		} else {
			s.logger.Debug("sleeping between log extractions", zap.Duration("duration", s.updateInterval))
			time.Sleep(s.updateInterval)
		}
	}
}

// QueryBlockEvents queries the `blocksMap` for events associated with a specific block number.
func (s *Sidecar) QueryBlockEvents(ctx context.Context, blockNumber *big.Int) ([]sidecartypes.Event, error) {
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
			return nil, fmt.Errorf("block %d was pruned or never fetched", blockNumber)
		}

		// Check the number of peers on the Ethereum node if we are not in development mode (net_peerCount not available
		// on anvil). We need to error if the Ethereum node has zero peers as it means that it can't sync up with the
		// network.
		if !s.development {
			peerCount, err := s.ethClient.PeerCount(ctx)
			if err != nil {
				s.logger.Error("error when fetching peer count", zap.Error(err))
				return nil, errors.New("could not get number of peers from node")
			}
			if peerCount == 0 {
				return nil, errors.New("detected zero peers; Ethereum node is not connected to the network")
			}
		}

		// Check if the Ethereum node is synced.
		syncProgress, err := s.ethClient.SyncProgress(ctx)
		if err != nil {
			return nil, fmt.Errorf("could not get syncing status from Ethereum node: %s", err.Error())
		}

		ethHeight, err := s.ethClient.BlockNumber(ctx)
		if err != nil {
			return nil, fmt.Errorf("could not get latest height from Ethereum node: %s", err.Error())
		}

		// If the sidecar is synced with Ethereum, and it processed the current Ethereum height already, then it must
		// be that the height being queried does not exist yet.
		isEthereumNodeSynced := syncProgress == nil
		nextEthereumBlock := new(big.Int).SetUint64(ethHeight + 1)
		sidecarSyncedWithEthereum := nextQueryBlock.Cmp(nextEthereumBlock) == 0
		if isEthereumNodeSynced && sidecarSyncedWithEthereum {
			return nil, fmt.Errorf("%s %s", sidecartypes.ErrBlockDoesNotExist, blockNumber)
		}

		// Compute the highest known height of the network.
		var networkHeight *big.Int
		if !isEthereumNodeSynced {

			// If the Ethereum node is not synced, then get the alleged network height from the SyncProgress object.
			networkHeight = new(big.Int).SetUint64(syncProgress.HighestBlock)
		} else {

			// If the Ethereum node is synced, then the network height is equivalent to the last synced height
			networkHeight = new(big.Int).SetUint64(ethHeight)
		}

		// Sidecar height is the height of the sidecar's next block to query minus 1
		sidecarHeight := new(big.Int).Sub(nextQueryBlock, big.NewInt(1))

		// If the delay is acceptable, return a special error for possibly different handling in the Sequencer.
		delay := new(big.Int).Sub(networkHeight, sidecarHeight)
		threshold := new(big.Int).SetUint64(s.acceptableDelay)
		if delay.Cmp(threshold) <= 0 {
			return nil, fmt.Errorf(
				"%s; Sidecar height %s, Ethereum height %s",
				sidecartypes.ErrSidecarFallenBehindWithAcceptableDelay,
				sidecarHeight.String(),
				networkHeight.String(),
			)
		}

		// Otherwise this block was not yet processed
		return nil, fmt.Errorf("block not yet processed %s", blockNumber)
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
