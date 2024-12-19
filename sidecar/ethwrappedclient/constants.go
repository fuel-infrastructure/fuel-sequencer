package ethwrappedclient

import "time"

const (
	BlockTime time.Duration = 12 * time.Second

	// Duration after which sidecar considers header subscription to be stale.
	// Reason: (Eth block time duration + 1s buffer) * (2 for generous window)
	HeaderSyncTimeout time.Duration = (BlockTime + time.Second) * 2
)
