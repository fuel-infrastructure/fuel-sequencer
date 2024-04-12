package store

import (
	"math/big"
	"strconv"
	"sync"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	"go.uber.org/zap"
)

// EventStore holds and manages Ethereum events.
type EventStore struct {
	mu        sync.Mutex
	blocksMap map[string]*sidecartypes.EthereumBlock

	// startQueryBlock is the block at which we started querying for events.
	// It also indicates the first block that we have in state.
	// If startQueryBlock >= nextQueryBlock then we have no blocks in state.
	startQueryBlock *big.Int

	// nextQueryBlock is the next block to be queried for events.
	// It also points to the block right after the latest one in state.
	nextQueryBlock *big.Int

	// blockPruneBuffer is the number of blocks of events to keep even if state is pruned.
	// Example: if we know the Sequencer only needs from block 100, we still keep block 90+.
	blockPruneBuffer uint64

	// maxQueryRange is the maximum number of Ethereum blocks per query.
	maxQueryRange *big.Int
}

// NewEventStore creates a new EventStore instance.
func NewEventStore(
	startQueryBlock *big.Int,
	nextQueryBlock *big.Int,
	maxQueryRange *big.Int,
) *EventStore {
	return &EventStore{
		blocksMap:        make(map[string]*sidecartypes.EthereumBlock),
		startQueryBlock:  startQueryBlock,
		nextQueryBlock:   nextQueryBlock,
		maxQueryRange:    maxQueryRange,
		blockPruneBuffer: 10,
	}
}

// AddEvents adds a map of events to the store.
func (store *EventStore) AddEvents(eventsMap *map[uint64][]sidecartypes.Event) {
	store.mu.Lock()
	defer store.mu.Unlock()

	for blockNumber, events := range *eventsMap {

		blockNumStr := strconv.FormatUint(blockNumber, 10)

		if block, exists := store.blocksMap[blockNumStr]; exists {
			// If block exists, append the new events to the existing slice
			block.Events = append(block.Events, events...)
		} else {
			// If block does not exist, create a new block and set its events
			store.blocksMap[blockNumStr] = &sidecartypes.EthereumBlock{
				BlockNumber: new(big.Int).SetUint64(blockNumber),
				Events:      events,
			}
		}
	}
}

// QueryEvents returns events for a given block number.
func (store *EventStore) QueryEvents(blockNumber *big.Int) ([]sidecartypes.Event, bool) {
	store.mu.Lock()
	defer store.mu.Unlock()

	block, exists := store.blocksMap[blockNumber.String()]
	if !exists {
		return nil, false
	}

	// Create a deep copy of the events to return.
	events := make([]sidecartypes.Event, len(block.Events))
	copy(events, block.Events)

	return events, true
}

// GetStartQueryBlock returns the startQueryBlock safely.
func (store *EventStore) GetStartQueryBlock() *big.Int {
	store.mu.Lock()
	defer store.mu.Unlock()

	// Return a copy to avoid external modification.
	if store.startQueryBlock == nil {
		return nil
	}
	return new(big.Int).Set(store.startQueryBlock)
}

// SetNextQueryBlock safely sets the value of nextQueryBlock.
func (store *EventStore) SetNextQueryBlock(blockNumber *big.Int) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.nextQueryBlock = new(big.Int).Set(blockNumber)
}

// GetNextQueryBlock returns the nextQueryBlock safely.
func (store *EventStore) GetNextQueryBlock() *big.Int {
	store.mu.Lock()
	defer store.mu.Unlock()

	// Return a copy to avoid external modification.
	if store.nextQueryBlock == nil {
		return nil
	}
	return new(big.Int).Set(store.nextQueryBlock)
}

// GetMaxQueryRange returns the nextQueryBlock safely.
func (store *EventStore) GetMaxQueryRange() *big.Int {
	store.mu.Lock()
	defer store.mu.Unlock()

	// Return a copy to avoid external modification.
	if store.maxQueryRange == nil {
		return nil
	}
	return new(big.Int).Set(store.maxQueryRange)
}

// CalibrateBlocksAndPruneLogs narrows down the startQueryBlock and nextQueryBlock so that we avoid
// storing logs unnecessarily, and we fast-forward the sidecar if it's lagging behind for any reason.
func (store *EventStore) CalibrateBlocksAndPruneLogs(logger *zap.Logger, lastSyncedBlock *big.Int) {
	store.mu.Lock()
	defer store.mu.Unlock()

	if lastSyncedBlock == nil {
		return
	}

	// The next block to be queried should be greater than the last synced block.
	// If the Sequencer is ahead, fast-forward the next query block to the last synced block + 1.
	//
	// Example: if last synced is 100 and next query is 90, next query will become 101.
	// Example: if last synced is 100 and next query is 100, next query will become 101.
	if lastSyncedBlock.Cmp(store.nextQueryBlock) >= 0 {
		store.nextQueryBlock = new(big.Int).Add(lastSyncedBlock, big.NewInt(1))
	}

	// If we're storing blocks that the Sequencer does not need, update the start query block
	// since the Sequencer will never ask for the older blocks again. We also prune the state.
	// A prune buffer ensures that we keep data from some old Ethereum blocks, just in case.
	//
	// Example: if last synced is 100 and start query is 90, we can prune until 100 and set start query to 101.
	// Example: if last synced is 100 and start query is 100, we can prune until 100 and set start query to 101.
	if lastSyncedBlock.Cmp(store.startQueryBlock) >= 0 && lastSyncedBlock.Uint64() > store.blockPruneBuffer {
		pruneFrom := store.startQueryBlock.Uint64()
		pruneUntil := lastSyncedBlock.Uint64() - store.blockPruneBuffer

		// Warning: the above uint64 subtraction is only safe because we know lastSyncedBlock >= blockPruneBuffer

		// Prune if there's anything to prune
		if len(store.blocksMap) > 0 && pruneUntil > pruneFrom {

			logger.Info("Pruning state",
				zap.Uint64("from_block", pruneFrom),
				zap.Uint64("to_block", pruneUntil),
			)
			for i := pruneFrom; i <= pruneUntil; i++ {
				delete(store.blocksMap, strconv.FormatUint(i, 10))
			}

			store.startQueryBlock = new(big.Int).SetUint64(pruneUntil + 1)
		}
	}
}
