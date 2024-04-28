package sidecar

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync/atomic"

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
		development:     development,
		acceptableDelay: acceptableDelay,
	}
}

// StartFetching begins the sidecar's operation.
func (s *Sidecar) StartFetching(ctx context.Context) error {
	s.logger.Info("starting sidecar")

	// Initial check to verify Ethereum client connectivity and log fetching capability.
	_, err := s.ethClient.FilterLogs(ctx, s.eventStore.GetStartQueryBlock(), s.eventStore.GetStartQueryBlock())
	if err != nil {
		s.logger.Error("failed to fetch logs for initial check", zap.Error(err))
		return err
	}

	go s.queryAndStoreEvents(ctx)

	return nil
}

// queryAndStoreEvents continuously fetches logs from the Ethereum blockchain and processes them.
func (s *Sidecar) queryAndStoreEvents(ctx context.Context) {
	s.stopped.Store(false)
	defer s.stopped.Store(true)

	ch := make(chan *ethereumtypes.Header)

	s.logger.Info("subscribing to new headers", zap.Uint64("start_block", s.eventStore.GetStartQueryBlock().Uint64()))
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
			s.logger.Error("observed error from subscription to logs", zap.Error(err))
			// TODO: are there cases where we want to recreate the subscription here?
		case header := <-ch:

			// If the sidecar has been stopped exit.
			if s.IsStopped() {
				return
			}

			s.logger.Info("detected block", zap.Uint64("block", header.Number.Uint64()))

			// Fetch the last synced Ethereum block before querying for new logs
			lastSyncedBlock, err := s.sequencerClient.FetchLastEthereumBlockSynced(ctx)
			if err != nil {
				// Log the error if the last synced Ethereum block is not obtained.
				// Note; We should still attempt to process Ethereum blocks. Reason being is that if the processing
				// is skipped the Sequencer will not be able to produce the first block and the sidecar would not be
				// able to query the Sequencer, causing a deadlock.
				s.logger.Error("failed to obtain LastEthereumBlockSynced from Sequencer", zap.Error(err))
			} else {
				s.logger.Debug(
					"queried LastEthereumBlockSynced from Sequencer",
					zap.String("last_ethereum_block_synced", lastSyncedBlock.String()),
				)
			}

			// Prune any old events that are no longer necessary to keep.
			s.eventStore.PruneLogs(s.logger, lastSyncedBlock)

			// Process and store log.
			event, err := s.ethClient.ProcessLog(log)
			s.eventStore.AddEvent(log.BlockNumber, event)

			// TODO: add concept of block finalization to determine which block number we've extracted up until, and what can be returned to clients.
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
			return nil, errors.New("could not get syncing status from Ethereum node")
		}

		// If the Ethereum node is synced, then it must be that the height being queried does not exist yet.
		isEthereumNodeSynced := syncProgress == nil
		if isEthereumNodeSynced {
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
