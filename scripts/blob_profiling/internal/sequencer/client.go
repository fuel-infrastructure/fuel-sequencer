package sequencer

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/x/evidence"
	"cosmossdk.io/x/upgrade"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"

	"github.com/fuel-infrastructure/fuel-sequencer/app"
	blobmodule "github.com/fuel-infrastructure/fuel-sequencer/x/blob/module"
	bridgemodule "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/module"
	sequencingmodule "github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/module"
)

const (
	chainID = "fuelsequencer-1" // Change to your chain ID
	denom   = "ufuel"           // Token denomination
)

type Client struct {
	rpcURL  string
	Sender  Account
	topic   string
	timeout time.Duration // Timeout till blob metadata is finalised

	keyName   string
	clientCtx client.Context
	txFactory tx.Factory
	keyring   keyring.Keyring
}

type EncodingConfig = testutil.TestEncodingConfig

// makeEncodingConfig creates the encoding configuration with all required module registrations
func makeEncodingConfig() EncodingConfig {
	modules := []module.AppModuleBasic{
		auth.AppModuleBasic{},
		authzmodule.AppModuleBasic{},
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
		distribution.AppModuleBasic{},
		consensus.AppModuleBasic{},
		slashing.AppModuleBasic{},
		mint.AppModuleBasic{},
		gov.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		bridgemodule.AppModuleBasic{},
		sequencingmodule.AppModuleBasic{},
		blobmodule.AppModuleBasic{},
	}
	return testutil.MakeTestEncodingConfig(modules...)
}

func NewClient(
	ctx context.Context, rpcURL, topic, sender string, blobTimeout time.Duration) (
	*Client, error,
) {
	// Initialize RPC client
	rpcClient, err := rpchttp.New(rpcURL, "/websocket")
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC client with URL %s: %w", rpcURL, err)
	}

	// Initialize SDK configuration with correct address prefix
	app.InitSDKConfig()

	// Create client context
	encodingConfig := makeEncodingConfig()
	clientCtx := client.Context{}.
		WithCodec(encodingConfig.Codec).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(nil).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastSync).
		WithClient(rpcClient).
		WithChainID(chainID)

	// Create keyring
	kr := keyring.NewInMemory(encodingConfig.Codec)

	// Create acc in keyring from mnemonic
	acc, exists := Accounts[sender]
	if !exists {
		return nil, fmt.Errorf("sender account %s not found", sender)
	}

	mnemonic := acc.Mnemonic
	hdPath := hd.CreateHDPath(118, 0, 0) // Cosmos HD path
	keyName := "sender"

	// Use the in-memory keyring to store the key
	keyRecord, err := kr.NewAccount(keyName, mnemonic, "", hdPath.String(), hd.Secp256k1)
	if err != nil {
		return nil, fmt.Errorf("failed to create account in keyring: %w", err)
	}

	// Get the address from the key record
	addr, err := keyRecord.GetAddress()
	if err != nil {
		return nil, fmt.Errorf("failed to get address from key record: %w", err)
	}

	// Update client context with keyring and from address info
	clientCtx = clientCtx.
		WithNodeURI(rpcURL).
		WithKeyring(kr).
		WithFromName(keyName).
		WithFromAddress(addr).
		WithSkipConfirmation(true)

	// Create a modified account with the actual derived address
	senderAccount := Account{
		Name:     acc.Name,
		Address:  addr.String(), // Use the address derived from the mnemonic
		Mnemonic: acc.Mnemonic,
	}

	// Query account to get existing sequence
	account, err := clientCtx.AccountRetriever.GetAccount(clientCtx, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	senderAccount.Sequence = account.GetSequence()

	return &Client{
		rpcURL:    rpcURL,
		Sender:    senderAccount,
		topic:     topic,
		keyName:   keyName,
		timeout:   blobTimeout,
		clientCtx: clientCtx,
		txFactory: tx.Factory{},
		keyring:   kr,
	}, nil
}
