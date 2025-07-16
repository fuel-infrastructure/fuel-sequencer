package keeper

import (
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	blobstore "github.com/fuel-infrastructure/blob-storage/pkg/store"
	"github.com/fuel-infrastructure/fuel-sequencer/x/blob/types"
)

// Package keeper implements the blob module keeper that integrates with the existing blob-storage package.
//
// The keeper supports multiple storage backends:
// - Mock implementation for testing (default)
// - Redis implementation for production (via SetBlobTxStore)
//
// Example usage for production with Redis:
//
//	import "github.com/fuel-infrastructure/blob-storage/pkg/store/redis"
//
//	redisStore, err := redis.NewRedisStore(ctx, "localhost:6379", "", 0, logger.Console)
//	if err != nil {
//		panic(err)
//	}
//	keeper.SetBlobTxStore(redisStore)
//

// Keeper is the blob module keeper
type Keeper struct {
	cdc          codec.BinaryCodec
	storeService store.KVStoreService
	logger       log.Logger
	tStoreKey    storetypes.StoreKey // Transient store key for blob data during block execution

	// Services for blob operations
	blobStoreService blobstore.BlobStore

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string
}

// NewKeeper creates a new blob Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	logger log.Logger,
	tStoreKey storetypes.StoreKey,
	authority string,
) *Keeper {
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}

	return &Keeper{
		cdc:          cdc,
		storeService: storeService,
		logger:       logger,
		tStoreKey:    tStoreKey,
		authority:    authority,
	}
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger() log.Logger {
	return k.logger.With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
