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

	// blockPruneBuffer is the number of blocks of events to keep even if state is pruned.
	// Example: if we know the Sequencer only needs from block 100, we still keep block 90+.
	blockPruneBuffer uint64
}

// NewEventStore creates a new EventStore instance.
func NewEventStore(startQueryBlock *big.Int) *EventStore {
	return &EventStore{
		blocksMap:        make(map[uint64]*sidecartypes.EthereumBlock),
		startQueryBlock:  startQueryBlock,
		blockPruneBuffer: 10,
	}
}

// AddEvent adds a new event to the store at the specified block.
func (store *EventStore) AddEvent(blockNumber uint64, event *sidecartypes.Event) {
	store.mu.Lock()
	defer store.mu.Unlock()

	block, exists := store.blocksMap[blockNumber]
	if exists {
		// If an entry exists for the block, append the new events to the existing events slice.
		block.Events = append(block.Events, *event)
	} else {
		// If an entry does not exist for the block, create a new entry and initialise with one event.
		store.blocksMap[blockNumber] = &sidecartypes.EthereumBlock{
			BlockNumber: new(big.Int).SetUint64(blockNumber),
			Events:      []sidecartypes.Event{*event},
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

// PruneLogs prunes state based on the last synced block so that we avoid storing logs unnecessarily.
// If pruning takes place, startQueryBlock is updated to reflect the first (i.e. oldest) block we have in state.
func (store *EventStore) PruneLogs(logger *zap.Logger, lastSyncedBlock *big.Int) {
	store.mu.Lock()
	defer store.mu.Unlock()

	// Last synced block is nil do not continue.
	if lastSyncedBlock == nil {
		return
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
