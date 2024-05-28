package store

import (
	"math/big"
	"sync"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	"go.uber.org/zap"
)

// EventStore holds and manages Ethereum events.
type EventStore struct {
	mu        sync.Mutex
	blocksMap map[uint64]*sidecartypes.EthereumBlock

	// startQueryBlock is the block at which we started querying for events.
	// It also indicates the oldest block that we might have in state.
	startQueryBlock *big.Int

	// endQueryBlock is the last block that we will query for events.
	endQueryBlock *big.Int

	// lastSyncedBlock is the last block queried for events.
	lastSyncedBlock *big.Int

	// blockPruneBuffer is the number of blocks of events to keep even if state is pruned.
	// Example: if we know the Sequencer only needs from block 100, we still keep block 90+.
	blockPruneBuffer uint64

	// maxQueryRange is the maximum number of Ethereum blocks per query.
	maxQueryRange *big.Int
}

// NewEventStore creates a new EventStore instance.
func NewEventStore(startQueryBlock, endQueryBlock, maxQueryRange *big.Int) *EventStore {
	return &EventStore{
		blocksMap:       make(map[uint64]*sidecartypes.EthereumBlock),
		startQueryBlock: startQueryBlock,
		endQueryBlock:   endQueryBlock, // might be nil
		// The last synced block is the one right before the one we're starting at.
		// The next query block will then evaluate to the last synced block + 1.
		lastSyncedBlock:  new(big.Int).Sub(startQueryBlock, big.NewInt(1)),
		maxQueryRange:    maxQueryRange,
		blockPruneBuffer: 10,
	}
}

// AddEvents adds a map of events to the store.
func (store *EventStore) AddEvents(eventsMap map[uint64][]sidecartypes.Event) {
	store.mu.Lock()
	defer store.mu.Unlock()

	for blockNumber, events := range eventsMap {

		if block, exists := store.blocksMap[blockNumber]; exists {
			// If an entry exists, append the new events to the existing slice.
			block.Events = append(block.Events, events...)
		} else {
			// If no entry was found, initialise a new one.
			store.blocksMap[blockNumber] = &sidecartypes.EthereumBlock{
				BlockNumber: new(big.Int).SetUint64(blockNumber),
				Events:      events,
			}
		}
	}
}

// GetStoredEvents returns events for a given block number.
func (store *EventStore) GetStoredEvents(blockNumber *big.Int) ([]sidecartypes.Event, bool) {
	store.mu.Lock()
	defer store.mu.Unlock()

	block, exists := store.blocksMap[blockNumber.Uint64()]
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

// GetEndQueryBlock returns the endQueryBlock safely.
func (store *EventStore) GetEndQueryBlock() *big.Int {
	store.mu.Lock()
	defer store.mu.Unlock()

	// Return a copy to avoid external modification.
	if store.endQueryBlock == nil {
		return nil
	}
	return new(big.Int).Set(store.endQueryBlock)
}

// SetLastSyncedBlock safely sets the value of lastSyncedBlock.
func (store *EventStore) SetLastSyncedBlock(blockNumber *big.Int) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.lastSyncedBlock = new(big.Int).Set(blockNumber)
}

// GetLastSyncedBlock returns the lastSyncedBlock safely.
func (store *EventStore) GetLastSyncedBlock() *big.Int {
	store.mu.Lock()
	defer store.mu.Unlock()

	// Return a copy to avoid external modification.
	if store.lastSyncedBlock == nil {
		return nil
	}
	return new(big.Int).Set(store.lastSyncedBlock)
}

// GetNextQueryBlock returns the next block to be queries, i.e. the last synced block + 1.
func (store *EventStore) GetNextQueryBlock() *big.Int {
	return new(big.Int).Add(store.GetLastSyncedBlock(), big.NewInt(1))
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

// CalibrateBlocksAndPruneLogs prunes state based on the last block synced by the Sequencer, so that we avoid storing
// logs unnecessarily. If pruning takes place, startQueryBlock is updated to reflect the first (i.e. oldest) block we
// have in state. Additionally, if the Sequencer is ahead, fast-forward the lastSyncedBlock to match the Sequencer.
func (store *EventStore) CalibrateBlocksAndPruneLogs(logger *zap.Logger, lastSyncedBlock *big.Int) {
	store.mu.Lock()
	defer store.mu.Unlock()

	// Last synced block is nil do not continue.
	if lastSyncedBlock == nil {
		return
	}

	// If the Sequencer is ahead, fast-forward the last synced block to that of the Sequencer.
	// Example: if last synced of the Sequencer is 100, last synced of the Sidecar will become 100.
	if lastSyncedBlock.Cmp(store.lastSyncedBlock) > 0 {
		store.lastSyncedBlock = lastSyncedBlock
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

			logger.Info("pruning state",
				zap.Uint64("from_block", pruneFrom),
				zap.Uint64("to_block", pruneUntil),
			)
			for i := pruneFrom; i <= pruneUntil; i++ {
				delete(store.blocksMap, i)
			}

			store.startQueryBlock = new(big.Int).SetUint64(pruneUntil + 1)
		}
	}
}
