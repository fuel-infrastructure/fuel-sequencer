package sidecar

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync/atomic"
	"time"

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

	updateInterval time.Duration

	// stopped indicates if the main process of the sidecar has been stopped or not.
	stopped atomic.Bool
}

// NewSidecar initializes a new Sidecar instance.
func NewSidecar(
	logger *zap.Logger,
	ethClient *ethclient.EthWrappedClient,
	sequencerClient *sequencerclient.SequencerClient,
	eventStore *store.EventStore,
) *Sidecar {
	return &Sidecar{
		logger:          logger,
		ethClient:       ethClient,
		sequencerClient: sequencerClient,
		eventStore:      eventStore,
		updateInterval:  10 * time.Second,
	}
}

// StartFetching begins the sidecar's operation.
func (s *Sidecar) StartFetching(ctx context.Context) error {
	s.logger.Info("starting sidecar")

	// Initial check to verify Ethereum client connectivity and log fetching capability.
	_, err := s.ethClient.FilterLogs(ctx, s.eventStore.GetStartQueryBlock(), s.eventStore.GetStartQueryBlock())
	if err != nil {
		s.logger.Error("Failed to fetch logs for initial check", zap.Error(err))
		return err
	}

	go s.queryAndStoreEvents(ctx)

	return nil
}

// queryAndStoreEvents continuously fetches logs from the Ethereum blockchain and processes them.
func (s *Sidecar) queryAndStoreEvents(ctx context.Context) {

	// The first block to be queried is the startQueryBlock.
	s.eventStore.SetNextQueryBlock(s.eventStore.GetStartQueryBlock())

	s.stopped.Store(false)
	defer s.stopped.Store(true)

	ticker := time.NewTicker(s.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.ShutDown()
			s.logger.Info("sidecar stopped via context")
		case <-ticker.C:

			// If the sidecar has been stopped exit.
			if s.IsStopped() {
				return
			}

			s.logger.Info("Processing from block", zap.Uint64("block", s.eventStore.GetNextQueryBlock().Uint64()))

			// Fetch the last synced Ethereum block before querying for new logs
			lastSyncedBlock, err := s.sequencerClient.FetchLastSyncedEthereumBlock(ctx)
			s.logger.Debug("Querying LastSyncedEthereumBlock: ", zap.String("last synced block", lastSyncedBlock.String()))
			if err != nil {
				// Log the error if the last synced Ethereum block is not obtained.
				// Note; We should still attempt to process Ethereum blocks. Reason being is that if the processing
				// is skipped the Sequencer will not be able to produce the first block and the sidecar would not be
				// able to query the Sequencer, causing a deadlock.
				s.logger.Error("failed to obtain last synced block from Sequencer", zap.Error(err))
			}

			// Prune any old events that are no longer necessary to keep.
			s.eventStore.CalibrateBlocksAndPruneLogs(s.logger, lastSyncedBlock)

			// Fetch and process logs based the next query block and the max range.
			eventsMap, nextQueryBlock := s.ethClient.FetchAndProcessLogs(
				ctx, s.logger, s.eventStore.GetNextQueryBlock(), s.eventStore.GetMaxQueryRange(),
			)

			// Store the newly fetched events and logs if they exist
			if eventsMap != nil {
				s.eventStore.AddEvents(eventsMap)
			}

			// Update the next query block
			s.eventStore.SetNextQueryBlock(nextQueryBlock)
		}
	}
}

// QueryBlockEvents queries the `blocksMap` for events associated with a specific block number.
func (s *Sidecar) QueryBlockEvents(ctx context.Context, blockNumber *big.Int) ([]sidecartypes.Event, error) {
	s.logger.Debug("Querying Block Events ", zap.String("block_number", blockNumber.String()))

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

		// Check if the Ethereum node is synced.
		isEthereumNodeSynced, err := s.ethClient.CheckEthereumNodeSync(ctx)
		if err != nil {
			return nil, err
		}

		ethHeight, err := s.ethClient.BlockNumber(ctx)
		if err != nil {
			return nil, errors.New("could not get latest height from Ethereum node")
		}

		// If the sidecar is synced with Ethereum, and it processed the current Ethereum height already, then it must
		// be that the height being queried does not exist yet.
		nextEthereumBlock := new(big.Int).SetUint64(ethHeight + 1)
		sidecarSyncedWithEthereum := nextQueryBlock.Cmp(nextEthereumBlock) == 0
		if isEthereumNodeSynced && sidecarSyncedWithEthereum {
			return nil, fmt.Errorf("%s %s", sidecartypes.ErrBlockDoesNotExist, blockNumber)
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
