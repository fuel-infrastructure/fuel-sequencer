package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/fuel-infrastructure/fuel-sequencer/x/blob/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService store.KVStoreService
		logger       log.Logger

		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string

		// Configuration
		blobhubAddress       string
		blobpoolRedisAddress string

		initialised bool           // initialise blobhub and blobpool connections
		*Blobpool                  // node storage for unconfirmed blob transactions
		blobhub     *blobhubClient // client for syncing with blobhub
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	logger log.Logger,
	authority string,
) Keeper {
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}

	return Keeper{
		cdc:                  cdc,
		storeService:         storeService,
		authority:            authority,
		logger:               logger,
		blobhubAddress:       "localhost:31035", // default value
		blobpoolRedisAddress: "localhost:6380",  // default value
		initialised:          false,             // Don't initialize yet
	}
}

// Initialize initializes the keeper's connections and services
// This should be called only when the app is actually running
func (k *Keeper) Initialize(ctx context.Context) error {
	if k.initialised {
		return nil // Already initialized
	}

	blobpool, err := newBlobpool(ctx, k.logger, k.blobpoolRedisAddress)
	if err != nil {
		return err
	}
	blobhubClient, err := newBlobhubClient(ctx, k.logger, blobpool, k.blobhubAddress)
	if err != nil {
		return err
	}

	// Blobpool Server exists to query blobpool, mainly used by the profiler
	// Server performance characteristics have not yet been identified.
	// To avoid unnecessary bottlenecks:
	// opting to disable blobpool server, and corresponding profiler measures.
	// go func() {
	// 	sl := k.logger.With("server")
	// 	sl.Info("starting blobpool server", "address", BlobpoolAddress)
	// 	err := k.server.StartServer(BlobpoolAddress)
	// 	if err != nil {
	// 		sl.Error("blobpool server failed", "error", err)
	// 		panic(err)
	// 	}
	// }()

	k.Blobpool = blobpool
	k.blobhub = blobhubClient
	k.initialised = true

	k.logger.Info("blob keeper initialized")
	return nil
}

// SetBlobhubAddress sets the blobhub address for the keeper
func (k *Keeper) SetBlobhubAddress(address string) {
	k.blobhubAddress = address
}

// SetBlobpoolRedisAddress sets the blobpool redis address for the keeper
func (k *Keeper) SetBlobpoolRedisAddress(address string) {
	k.blobpoolRedisAddress = address
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger() log.Logger {
	return k.logger.With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
