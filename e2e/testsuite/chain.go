package testsuite

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"cosmossdk.io/core/address"
	"cosmossdk.io/x/evidence"
	"cosmossdk.io/x/upgrade"
	cmtbytes "github.com/cometbft/cometbft/libs/bytes"
	cmrand "github.com/cometbft/cometbft/libs/rand"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkAddressCodec "github.com/cosmos/cosmos-sdk/codec/address"
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
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fuel-infrastructure/fuel-sequencer/app"
	appcodec "github.com/fuel-infrastructure/fuel-sequencer/app/codec"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	bridge "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/module"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	sequencing "github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/module"
	grpcencoding "google.golang.org/grpc/encoding"
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
	Cdc            codec.Codec
	TestCdc        codec.Codec // an exported alias of cdc
	TestGrpcCdc    grpcencoding.Codec

	addressCdc     address.Codec
	TestAddressCdc address.Codec // an exported alias of cdc
)

func init() {
	modules := []module.AppModuleBasic{
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
		distribution.AppModuleBasic{},
		consensus.AppModuleBasic{},
		slashing.AppModuleBasic{},
		mint.AppModuleBasic{},
		gov.AppModuleBasic{},
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
	encodingConfig.InterfaceRegistry.RegisterImplementations(
		(*sdk.AccountI)(nil),
		&bridgetypes.EthOwnedBaseAccount{},
		&bridgetypes.EthOwnedContinuousVestingAccount{},
		&vestingtypes.DelayedVestingAccount{},
		&vestingtypes.ContinuousVestingAccount{},
	)

	Cdc = encodingConfig.Codec
	TestCdc = Cdc
	TestGrpcCdc = Cdc.(codec.GRPCCodecProvider).GRPCCodec()

	addressCdc = appcodec.NewFuelSequencerAddressCodec(sdkAddressCodec.NewBech32Codec(app.AccountAddressPrefix))
	TestAddressCdc = addressCdc
}

type Chain struct {
	DataDir    string
	id         string
	numNodes   int
	Validators []*validator

	grpcClients   *GRPCClients
	rpcClient     *rpchttp.HTTP
	sidecarClient sidecartypes.SidecarClient
	ethClient     *ethclient.Client
}

func NewFixedChain(chainName, dataDir string, numNodes int) (*Chain, error) {
	return &Chain{
		id:       chainName,
		DataDir:  dataDir,
		numNodes: numNodes,
	}, nil
}

func newChain(numNodes int) (*Chain, error) {
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

	return &Chain{
		id:       "Chain-" + cmrand.NewRand().Str(6),
		DataDir:  tmpDir,
		numNodes: numNodes,
	}, nil
}

func (c *Chain) ConfigDir() string {
	return fmt.Sprintf("%s/%s", c.DataDir, c.id)
}

// createAndInitFuelSequencerValidators initialises FuelSequencer nodes with mnemonics (if specified) or random keys.
func (c *Chain) createAndInitFuelSequencerValidators(mnemonics []string) error {
	return CreateAndInitFuelSequencerValidatorsFromGenesis(c, mnemonics)
}

func CreateAndInitFuelSequencerValidatorsFromGenesis(c *Chain, mnemonics []string) error {
	// Determine whether to use mnemonics.
	useMnemonics := len(mnemonics) > 0

	for i := 0; i < c.numNodes; i++ {
		node := c.createFuelSequencerValidator(i)

		// generate genesis files
		if err := node.init(); err != nil {
			return err
		}

		c.Validators = append(c.Validators, node)

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

func (c *Chain) createFuelSequencerValidator(index int) *validator {
	return &validator{
		chain:   c,
		index:   index,
		Moniker: validatorMonikerPrefix,
	}
}

func (c *Chain) clientContext(
	nodeURI string, kb *keyring.Keyring, fromName string, fromAddr sdk.AccAddress, outputBuffer *bytes.Buffer,
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
		WithOutput(outputBuffer).
		WithFrom(fromName).
		WithFromName(fromName).
		WithFromAddress(fromAddr).
		WithSkipConfirmation(true)

	return &clientContext, nil
}

func (c *Chain) sendMsgs(
	clientCtx client.Context,
	outputBuffer *bytes.Buffer,
	gas uint64,
	msgs ...sdk.Msg,
) (*sdk.TxResponse, error) {

	txf := tx.Factory{}.
		WithAccountRetriever(clientCtx.AccountRetriever).
		WithChainID(c.id).
		WithTxConfig(clientCtx.TxConfig).
		WithGasAdjustment(1.2).
		WithKeybase(clientCtx.Keyring).
		WithGas(gas).
		WithGasPrices(fmt.Sprintf("%s%s", minGasPrices, BridgeDenom)).
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

	err := tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msgs...)
	if err != nil {
		return nil, err
	}

	err = WaitForCondition(time.Second*30, time.Millisecond*500, func() (bool, error) {
		return outputBuffer.Len() > 0, nil
	})
	resBytes := outputBuffer.Bytes()

	var res sdk.TxResponse
	err = Cdc.UnmarshalJSON(resBytes, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (c *Chain) SubscribeToSequencer(ctx context.Context, query string) (<-chan coretypes.ResultEvent, error) {
	res, err := c.rpcClient.Subscribe(ctx, "", query)
	if err != nil {
		return nil, fmt.Errorf("rpc client status: %w", err)
	}
	return res, nil
}

func (c *Chain) FuelSequencerHeight(ctx context.Context) (uint64, error) {
	res, err := c.rpcClient.Status(ctx)
	if err != nil {
		return 0, fmt.Errorf("rpc client status: %w", err)
	}
	return uint64(res.SyncInfo.LatestBlockHeight), nil
}

func (c *Chain) GetBlockHeaderHash(ctx context.Context, height int64) (cmtbytes.HexBytes, error) {
	res, err := c.rpcClient.Block(ctx, &height)
	if err != nil {
		return cmtbytes.HexBytes{}, fmt.Errorf("rpc client status: %w", err)
	}
	return res.BlockID.Hash, nil
}

func (c *Chain) EthereumHeight(ctx context.Context) (uint64, error) {
	res, err := c.ethClient.BlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("rpc client status: %w", err)
	}
	return res, nil
}
