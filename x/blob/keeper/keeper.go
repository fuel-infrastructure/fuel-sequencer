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
		blobhubAddress        string
		blobpoolSqlitePath    string
		blobpoolServerEnabled bool
		blobpoolServerAddress string

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
		cdc:                   cdc,
		storeService:          storeService,
		authority:             authority,
		logger:                logger,
		blobhubAddress:        "localhost:31035",    // default value
		blobpoolSqlitePath:    "./data/blobpool.db", // default value
		blobpoolServerEnabled: false,                // default value
		blobpoolServerAddress: "localhost:21025",    // default value
		initialised:           false,                // Don't initialize yet
	}
}

// Initialize initializes the keeper's connections and services
// This should be called only when the app is actually running
func (k *Keeper) Initialize(ctx context.Context) error {
	if k.initialised {
		return nil // Already initialized
	}

	blobpool, err := newBlobpool(ctx, k.logger, k.blobpoolSqlitePath, k.blobpoolServerEnabled)
	if err != nil {
		return err
	}
	blobhubClient, err := newBlobhubClient(ctx, k.logger, blobpool, k.blobhubAddress)
	if err != nil {
		return err
	}

	// Start blobpool server if enabled
	if k.blobpoolServerEnabled {
		go func() {
			sl := k.logger.With("server")
			sl.Info("starting blobpool server", "address", k.blobpoolServerAddress)
			err := blobpool.server.StartServer(k.blobpoolServerAddress)
			if err != nil {
				sl.Error("blobpool server failed", "error", err)
				panic(err)
			}
		}()
	}

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

// SetBlobpoolSqlitePath sets the blobpool sqlite path for the keeper
func (k *Keeper) SetBlobpoolSqlitePath(path string) {
	k.blobpoolSqlitePath = path
}

// SetBlobpoolServerEnabled sets whether the blobpool server should be enabled
func (k *Keeper) SetBlobpoolServerEnabled(enabled bool) {
	k.blobpoolServerEnabled = enabled
}

// SetBlobpoolServerAddress sets the blobpool server address for the keeper
func (k *Keeper) SetBlobpoolServerAddress(address string) {
	k.blobpoolServerAddress = address
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger() log.Logger {
	return k.logger.With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
