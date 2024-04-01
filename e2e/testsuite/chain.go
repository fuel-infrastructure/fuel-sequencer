package testsuite

import (
	"context"
	"fmt"
	"github.com/cometbft/cometbft/libs/bytes"
	"os"

	"cosmossdk.io/x/evidence"
	"cosmossdk.io/x/upgrade"
	cmtbytes "github.com/cometbft/cometbft/libs/bytes"
	cmrand "github.com/cometbft/cometbft/libs/rand"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/ethclient"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	bridge "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/module"
	sequencing "github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/module"
)

const (
	keyringPassphrase = "testpassphrase"
	keyringAppName    = "testnet"

	tmpDirPattern = "fuelsequencer-e2e-testnet"

	validatorKeyName       = "val"
	validatorMonikerPrefix = "fuelsequencer"
)

var (
	encodingConfig testutil.TestEncodingConfig
	cdc            codec.Codec
)

func init() {
	// TODO: is this the correct way?
	modules := []module.AppModuleBasic{
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
		distribution.AppModuleBasic{},
		consensus.AppModuleBasic{},
		slashing.AppModuleBasic{},
		mint.AppModuleBasic{},
		gov.AppModuleBasic{},
		crisis.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		bridge.AppModuleBasic{},
		sequencing.AppModuleBasic{},
	}
	encodingConfig = testutil.MakeTestEncodingConfig(modules...)

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

	grpcClients   *GRPCClients
	rpcClient     *rpchttp.HTTP
	sidecarClient sidecartypes.SidecarClient
	ethClient     *ethclient.Client
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

	tmpDir, err := os.MkdirTemp(dir, tmpDirPattern)
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
			if err := node.createKeyFromMnemonic(validatorKeyName, mnemonics[i], ""); err != nil {
				return err
			}
		} else {
			if err := node.createKey(validatorKeyName); err != nil {
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
		moniker: validatorMonikerPrefix,
	}
}

func (c *chain) clientContext(
	nodeURI string, kb *keyring.Keyring, fromName string, fromAddr sdk.AccAddress,
) (*client.Context, error) { //nolint:unparam

	// TODO: if anything goes wrong with unregistered types, might need to re-add some stuff to this function

	rpcClient, err := rpchttp.New(nodeURI, "/websocket")
	if err != nil {
		return nil, err
	}

	clientContext := client.Context{}.
		WithChainID(c.id).
		WithCodec(encodingConfig.Codec).
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

	// TODO: make customisable
	txf = txf.WithFees(fmt.Sprintf("246913560%s", BridgeDenom))

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

func (c *chain) FuelSequencerHeight(ctx context.Context) (uint64, error) {
	res, err := c.rpcClient.Status(ctx)
	if err != nil {
		return 0, fmt.Errorf("rpc client status: %w", err)
	}
	return uint64(res.SyncInfo.LatestBlockHeight), nil
}

func (c *chain) GetBlockHeaderHash(ctx context.Context, height int64) (cmtbytes.HexBytes, error) {
	res, err := c.rpcClient.Block(ctx, &height)
	if err != nil {
		return cmtbytes.HexBytes{}, fmt.Errorf("rpc client status: %w", err)
	}
	return res.BlockID.Hash, nil
}

func (c *chain) EthereumHeight(ctx context.Context) (uint64, error) {
	res, err := c.ethClient.BlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("rpc client status: %w", err)
	}
	return res, nil
}

func (c *chain) BridgeCommitment(ctx context.Context, start, end uint64) (bytes.HexBytes, error) {
	res, err := c.rpcClient.BridgeCommitment(ctx, start, end)
	if err != nil {
		return bytes.HexBytes{}, fmt.Errorf("rpc client status: %w", err)
	}
	return res.BridgeCommitment, nil
}
