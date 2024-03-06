package testsuite

import (
	"fmt"
	"os"

	cmrand "github.com/cometbft/cometbft/libs/rand"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkTypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth"
	sdkTx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	keyringPassphrase = "testpassphrase"
	keyringAppName    = "testnet"
)

var (
	encodingConfig testutil.TestEncodingConfig
	cdc            codec.Codec
)

func init() {
	// TODO: might need to use custom test encoding config to register our custom stuff
	bankModule := bank.AppModuleBasic{}
	authModule := auth.AppModuleBasic{}
	encodingConfig = testutil.MakeTestEncodingConfig(bankModule, authModule)

	encodingConfig.InterfaceRegistry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&stakingtypes.MsgCreateValidator{},
		&stakingtypes.MsgBeginRedelegate{},
	)
	encodingConfig.InterfaceRegistry.RegisterImplementations(
		(*cryptotypes.PubKey)(nil),
		&secp256k1.PubKey{},
		&ed25519.PubKey{},
	)

	cdc = encodingConfig.Codec
}

type chain struct {
	dataDir    string
	id         string
	numNodes   int
	validators []*validator
}

func newChain(numNodes int) (*chain, error) {
	var dir string
	var err error
	if _, found := os.LookupEnv("CI"); found {
		dir, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}

	tmpDir, err := os.MkdirTemp(dir, "fuelsequencer-e2e-testnet")
	if err != nil {
		return nil, err
	}

	return &chain{
		id:       "chain-" + cmrand.NewRand().Str(6),
		dataDir:  tmpDir,
		numNodes: numNodes,
	}, nil
}

func (c *chain) configDir() string {
	return fmt.Sprintf("%s/%s", c.dataDir, c.id)
}

// createAndInitFuelSequencerValidators initialises FuelSequencer nodes with mnemonics (if specified) or random keys.
func (c *chain) createAndInitFuelSequencerValidators(mnemonics []string) error {

	// Determine whether to use mnemonics.
	useMnemonics := len(mnemonics) > 0

	for i := 0; i < c.numNodes; i++ {
		node := c.createFuelSequencerValidator(i)

		// generate genesis files
		if err := node.init(); err != nil {
			return err
		}

		c.validators = append(c.validators, node)

		// create keys
		if useMnemonics {
			if err := node.createKey("val"); err != nil {
				return err
			}
		} else {
			if err := node.createKeyFromMnemonic("val", mnemonics[i], ""); err != nil {
				return err
			}
		}
		if err := node.createNodeKey(); err != nil {
			return err
		}
		if err := node.createConsensusKey(); err != nil {
			return err
		}
	}

	return nil
}

func (c *chain) createFuelSequencerValidator(index int) *validator {
	return &validator{
		chain:   c,
		index:   index,
		moniker: "fuelsequencer",
	}
}

func (c *chain) clientContext(nodeURI string, kb *keyring.Keyring, fromName string, fromAddr sdk.AccAddress) (*client.Context, error) { //nolint:unparam
	amino := codec.NewLegacyAmino()
	interfaceRegistry := sdkTypes.NewInterfaceRegistry()
	interfaceRegistry.RegisterImplementations((*sdk.Msg)(nil),
		&stakingtypes.MsgCreateValidator{},
	)
	interfaceRegistry.RegisterImplementations((*cryptotypes.PubKey)(nil), &secp256k1.PubKey{}, &ed25519.PubKey{})

	protoCodec := codec.NewProtoCodec(interfaceRegistry)
	txCfg := sdkTx.NewTxConfig(protoCodec, sdkTx.DefaultSignModes)

	encodingConfig := testutil.TestEncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             protoCodec,
		TxConfig:          txCfg,
		Amino:             amino,
	}
	// TODO: simapp.ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	// TODO: simapp.ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	rpcClient, err := rpchttp.New(nodeURI, "/websocket")
	if err != nil {
		return nil, err
	}

	clientContext := client.Context{}.
		WithChainID(c.id).
		WithCodec(protoCodec).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithNodeURI(nodeURI).
		WithClient(rpcClient).
		WithBroadcastMode(flags.BroadcastSync).
		WithKeyring(*kb).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithOutputFormat("json").
		WithFrom(fromName).
		WithFromName(fromName).
		WithFromAddress(fromAddr).
		WithSkipConfirmation(true)

	return &clientContext, nil
}

func (c *chain) sendMsgs(clientCtx client.Context, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	txf := tx.Factory{}.
		WithAccountRetriever(clientCtx.AccountRetriever).
		WithChainID(c.id).
		WithTxConfig(clientCtx.TxConfig).
		WithGasAdjustment(1.2).
		WithKeybase(clientCtx.Keyring).
		WithGas(12345678).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)

	fromAddr := clientCtx.GetFromAddress()

	if err := txf.AccountRetriever().EnsureExists(clientCtx, fromAddr); err != nil {
		return nil, err
	}

	initNum, initSeq := txf.AccountNumber(), txf.Sequence()
	if initNum == 0 || initSeq == 0 {
		num, seq, err := txf.AccountRetriever().GetAccountNumberSequence(clientCtx, fromAddr)
		if err != nil {
			return nil, err
		}

		if initNum == 0 {
			txf = txf.WithAccountNumber(num)
		}

		if initSeq == 0 {
			txf = txf.WithSequence(seq)
		}
	}

	txf = txf.WithFees("246913560ufuel")

	err := tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msgs...)
	if err != nil {
		return nil, err
	}

	resBytes := []byte{}
	_, err = clientCtx.Input.Read(resBytes)
	if err != nil {
		return nil, err
	}

	var res sdk.TxResponse
	err = cdc.Unmarshal(resBytes, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
