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
		cdc:          cdc,
		storeService: storeService,
		authority:    authority,
		logger:       logger,
		initialised:  false, // Don't initialize yet
	}
}

// Initialize initializes the keeper's connections and services
// This should be called only when the app is actually running
func (k *Keeper) Initialize(ctx context.Context) error {
	if k.initialised {
		return nil // Already initialized
	}

	blobpool := newBlobpool(ctx, k.logger)
	blobhubClient, err := newBlobhubClient(ctx, k.logger, blobpool)
	if err != nil {
		return err
	}

	go func() {
		sl := k.logger.With("server")
		sl.Info("starting blobpool server", "address", BlobpoolAddress)
		err := k.server.StartServer(BlobpoolAddress)
		if err != nil {
			sl.Error("blobpool server failed", "error", err)
			panic(err)
		}
	}()

	k.Blobpool = blobpool
	k.blobhub = blobhubClient
	k.initialised = true

	k.logger.Info("blob keeper initialized")
	return nil
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger() log.Logger {
	return k.logger.With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
