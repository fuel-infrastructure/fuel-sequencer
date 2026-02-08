package keeper

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"path/filepath"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cometbft/cometbft/privval"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	blobclient "github.com/fuel-infrastructure/blob-storage/pkg/client"

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
		chunkMode             bool // enable chunk-mode attestation
		chunkValidatorIndex   int  // which chunk this validator attests
		validatorID           string // derived from consensus key

		// Consensus key pair for Ed25519 attestation signing
		privKey ed25519.PrivateKey
		pubKey  ed25519.PublicKey

		// Staking keeper for on-chain DA attestation verification
		stakingKeeper types.StakingKeeper

		initialised bool           // initialise blobhub and blobpool connections
		*Blobpool                  // node storage for unconfirmed blob transactions
		blobhub     *blobhubClient // client for syncing with blobhub (handles both sync and ACK)
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
		validatorID:           "",                   // will be set from consensus key during Initialize
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

	// Create blobhub client config
	config := &blobclient.ClientConfig{
		BaseURL: normalizeBlobhubAddress(k.blobhubAddress),
		Logger:  nil, // Use no-op logger from client library
	}
	client := blobclient.NewClient(config)

	// Create blobhub client with validator ID and consensus key pair
	blobhubClient, err := newBlobhubClient(ctx, k.logger, blobpool, k.blobhubAddress, client, k.validatorID, k.chunkMode, k.chunkValidatorIndex, k.privKey, k.pubKey)
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

// SetChunkMode enables or disables chunk-mode attestation.
func (k *Keeper) SetChunkMode(enabled bool) {
	k.chunkMode = enabled
}

// SetChunkValidatorIndex sets the validator chunk index for chunk-mode attestation.
func (k *Keeper) SetChunkValidatorIndex(index int) {
	k.chunkValidatorIndex = index
}

// SetValidatorID sets the validator ID for the keeper (derived from consensus key)
func (k *Keeper) SetValidatorID(id string) {
	k.validatorID = id
}

// SetConsensusKeyPair sets the Ed25519 consensus key pair for attestation signing.
func (k *Keeper) SetConsensusKeyPair(priv ed25519.PrivateKey, pub ed25519.PublicKey) {
	k.privKey = priv
	k.pubKey = pub
}

// SetStakingKeeper sets the staking keeper for on-chain DA attestation verification.
func (k *Keeper) SetStakingKeeper(sk types.StakingKeeper) {
	k.stakingKeeper = sk
}

// LoadConsensusKeyPair loads the Ed25519 consensus key pair from priv_validator_key.json.
// CometBFT's ed25519.PrivKey is []byte (64 bytes: seed||pubkey), same format as Go stdlib.
func LoadConsensusKeyPair(nodeHome string) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config
	config.SetRoot(nodeHome)

	pvKeyFile := config.PrivValidatorKeyFile()
	pvStateFile := config.PrivValidatorStateFile()

	if _, err := filepath.Abs(pvKeyFile); err != nil {
		return nil, nil, fmt.Errorf("failed to resolve priv validator key file path: %w", err)
	}

	filePV := privval.LoadOrGenFilePV(pvKeyFile, pvStateFile)
	if filePV == nil {
		return nil, nil, fmt.Errorf("failed to load priv validator key")
	}

	// CometBFT's ed25519.PrivKey is just []byte — 64 bytes in seed||pubkey format.
	// filePV.Key.PrivKey.Bytes() returns the raw bytes directly.
	privBytes := filePV.Key.PrivKey.Bytes()
	if len(privBytes) != 64 {
		return nil, nil, fmt.Errorf("unexpected private key length: %d (expected 64)", len(privBytes))
	}

	priv := ed25519.PrivateKey(make([]byte, 64))
	copy(priv, privBytes)
	pub := priv.Public().(ed25519.PublicKey)

	return priv, pub, nil
}

// DeriveValidatorIDFromConsensusKey derives the validator ID from the node's consensus key.
// It loads the priv_validator_key.json file and extracts the public key address.
// Returns empty string if the key file doesn't exist (non-validator node).
func DeriveValidatorIDFromConsensusKey(nodeHome string) (string, error) {
	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config
	config.SetRoot(nodeHome)

	pvKeyFile := config.PrivValidatorKeyFile()
	pvStateFile := config.PrivValidatorStateFile()

	// Check if priv_validator_key.json exists
	if _, err := filepath.Abs(pvKeyFile); err != nil {
		return "", fmt.Errorf("failed to resolve priv validator key file path: %w", err)
	}

	filePV := privval.LoadOrGenFilePV(pvKeyFile, pvStateFile)
	if filePV == nil {
		return "", fmt.Errorf("failed to load priv validator key")
	}

	// Extract public key address and encode as hex
	pubKey, err := filePV.GetPubKey()
	if err != nil {
		return "", fmt.Errorf("failed to get public key from priv validator: %w", err)
	}
	if pubKey == nil {
		return "", fmt.Errorf("public key is nil")
	}

	address := pubKey.Address()
	validatorID := hex.EncodeToString(address)

	return validatorID, nil
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger() log.Logger {
	return k.logger.With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
