// Test suite inspired by https://github.com/PeggyJV/sommelier/tree/v7.0.1/integration_tests

package testsuite

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	osuser "os/user"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"cosmossdk.io/math"
	cmconfig "github.com/cometbft/cometbft/config"
	cmjson "github.com/cometbft/cometbft/libs/json"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cosmos/cosmos-sdk/server"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fuel-infrastructure/fuel-sequencer/app"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func init() {
	app.InitSDKConfig()
	app.InitCometBFTConfig()
	app.InitAppConfig()

	sdk.DefaultBondDenom = BridgeDenom
}

const (
	BridgeDenom       = "ufuel"
	minGasPrices      = "0.01"
	SupplyDeltaPeriod = uint64(10)

	// Balance and staked amount per validator
	initBalance = 210000000000 // per validator
	initStaked  = 100000000000 // per validator

	fuelSequencerDockerImageRepo = "fuel-infrastructure/fuel-sequencer"
	fuelSequencerDockerImageTag  = "latest"

	fuelSequencerValidatorDefaultHome = "/home/fuelsequencer/.fuelsequencer"
	fuelSequencerBinary               = "fuelsequencerd"

	ethereumDockerImageRepo = "fuel-infrastructure/contracts-docker-e2e"
	ethereumDockerImageTag  = "latest"

	succinctXOperatorDockerImageRepo = "fuel-infrastructure/fuel-stream-x-operator-docker-e2e"
	succinctXOperatorDockerImageTag  = "latest"

	succinctXRelayerDockerImageRepo = "fuel-infrastructure/fuel-stream-x-relayer-docker-e2e"
	succinctXRelayerDockerImageTag  = "latest"
)

var (
	// Balance and staked amount per validator
	InitBalanceCoin = sdk.NewInt64Coin(BridgeDenom, initBalance)
	InitStakedCoin  = sdk.NewInt64Coin(BridgeDenom, initStaked)

	// MNEMONICS dictates how many Sequencer nodes will be created by specifying their mnemonic.
	// The first mnemonic is reused for the Ethereum validator mnemonic.
	MNEMONICS = []string{
		// should match the one on test-contracts
		"test test test test test test test test test test test junk",
		// alice
		"dinner crash nurse casino baby fold race cheese elite column sausage sleep close royal rain over mechanic minimum outdoor conduct cash wagon frog evidence",
		// bob
		"gaze drama excess raven follow antenna swallow beef upper myself question pitch course ill adult century crisp ice rough match praise sing unveil vintage",
	}

	// ADDRESSES are the FuelSequencer addresses derived from the above MNEMONICS.
	ADDRESSES = []string{
		// first validator
		"fuelsequencer15yk64u7zc9g9k2yr2wmzeva5qgwxps6y3z4xeu",
		// alice
		"fuelsequencer1vtfzrk6f4m6kxt6ehyqt9j5su5hvcz5q3dmlsm",
		// bob
		"fuelsequencer163rsv65t4893t2rz5rmda9sly7lgdlq2jgr36m",
	}

	// FUEL_STREAM_X_CONTRACT is the FuelStreamX contract that generates events, deployed on the Ethereum node.
	FUEL_STREAM_X_CONTRACT = "0xa513E6E4b8f2a923D98304ec87F64353C4D5C853"
	// GATEWAY_CONTRACT is a contract by Succinct that does ZK proof verification.
	GATEWAY_CONTRACT = "0xDc64a140Aa3E981100a9becA4E685f962f0cF6C9"

	// Circuits
	NEXT_HEADER_FUNCTION_ID  = "0x6eed2eb6930917a14cc2c0732dfb6e4b8582a54d8c4fefae146a2261705e7a35"
	HEADER_RANGE_FUNCTION_ID = "0xd889dce37c711c10c083d156972c0cedf8d7da68a9db7e7ad63333c108815be8"
)

var (
	LogLevel = zaptest.Level(zap.DebugLevel)
)

type E2ETestSuite struct {
	suite.Suite

	log *zap.Logger

	chain         *chain
	dockerPool    *dockertest.Pool
	dockerNetwork *dockertest.Network

	// Ethereum
	ethResource *dockertest.Resource

	// Sequencer
	valResources []*dockertest.Resource

	// SuccinctX
	succinctOperatorResource *dockertest.Resource
	succinctRelayerResource  *dockertest.Resource
}

func (s *E2ETestSuite) SetupTest() {
	if testing.Short() {
		s.T().Skip()
	}

	s.T().Log("setting up E2E test...")

	s.log = zaptest.NewLogger(s.T(), LogLevel)

	var err error
	s.chain, err = newChain(len(MNEMONICS))
	s.Require().NoError(err)
	s.dockerPool, err = dockertest.NewPool("")
	s.Require().NoError(err)
	s.dockerNetwork, err = s.dockerPool.CreateNetwork(fmt.Sprintf("%s-testnet", s.chain.id))
	s.Require().NoError(err)

	s.T().Logf("starting E2E infrastructure; chain-id: %s; datadir: %s", s.chain.id, s.chain.dataDir)

	// initialization
	s.initFuelSequencerNodes(MNEMONICS)
	s.initEthereumNodes(MNEMONICS)

	// continue generating node genesis
	s.initFuelSequencerGenesis()
	s.initFuelSequencerValidatorConfigs()

	// container infrastructure
	s.runFuelSequencerValidators()

	// set up clients
	s.initGRPCClients()
	s.initRPCClient()
	s.initEthereumRPCClient()
	s.initSidecarClient()

	// We need the genesis header for solidity smart contracts
	err = s.WaitForBlocks(s.Ctx(), 1, time.Minute)
	s.Require().NoError(err)

	// Get genesis header
	block, err := s.GetBlockByHeight(s.Ctx(), 1)
	s.Require().NoError(err)
	block.Hash()
	s.Require().Equal(sequencerHeight, uint64(block.Header.Height))

	// run the eth container so that the contract addresses are available
	// TODO: probably run this after sequencer since we would need the genesis headers from the sequencer in the smart contracts
	s.runEthContainer()
}

func (s *E2ETestSuite) TearDownTest() {
	if str := os.Getenv("E2E_SKIP_CLEANUP"); len(str) > 0 {
		skipCleanup, err := strconv.ParseBool(str)
		s.Require().NoError(err)

		if skipCleanup {
			s.T().Log("skipping teardown")
			return
		}
	}

	s.T().Log("tearing down e2e integration test suite...")

	s.Require().NoError(os.RemoveAll(s.chain.dataDir))
	s.Require().NoError(s.dockerPool.Purge(s.ethResource))

	for _, vc := range s.valResources {
		s.Require().NoError(s.dockerPool.Purge(vc))
	}

	s.Require().NoError(s.dockerPool.RemoveNetwork(s.dockerNetwork))
}

// initFuelSequencerNodes initialises FuelSequencer nodes with mnemonics (if specified) or random keys.
// It also sets up the genesis file using the first validator and copies it to all other validator nodes.
func (s *E2ETestSuite) initFuelSequencerNodes(mnemonics []string) {
	s.Require().NoError(s.chain.createAndInitFuelSequencerValidators(mnemonics))

	// initialize a genesis file for the first validator
	val0ConfigDir := s.chain.validators[0].configDir()
	for _, val := range s.chain.validators {
		s.Require().NoError(
			addGenesisAccount(val0ConfigDir, "", InitBalanceCoin.String(), val.address()),
		)
	}

	// copy the genesis file to the remaining validators
	for _, val := range s.chain.validators[1:] {
		err := copyFile(
			filepath.Join(val0ConfigDir, "config", "genesis.json"),
			filepath.Join(val.configDir(), "config", "genesis.json"),
		)
		s.Require().NoError(err)
	}
}

// initEthereumNodes initialises Ethereum nodes with mnemonics or a node count (empty mnemonics list).
func (s *E2ETestSuite) initEthereumNodes(mnemonics []string) {
	// TODO: create genesis file instead of assuming it exists?

	// Determine whether to use mnemonics.
	useMnemonics := len(mnemonics) > 0

	for i, val := range s.chain.validators {
		if useMnemonics {
			s.Require().NoError(val.generateEthereumKeyFromMnemonic(mnemonics[i]))
		} else {
			s.Require().NoError(val.generateEthereumKey())
		}
	}
}

func (s *E2ETestSuite) initFuelSequencerGenesis() {
	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(s.chain.validators[0].configDir())
	config.Moniker = s.chain.validators[0].moniker

	genFilePath := config.GenesisFile()
	appGenState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFilePath)
	s.Require().NoError(err)

	var govGenState govtypesv1.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[govtypes.ModuleName], &govGenState))

	// set short voting period to allow gov proposals in tests
	seconds20 := time.Second * 20
	govGenState.Params.VotingPeriod = &seconds20
	govGenState.Params.MinDeposit = sdk.Coins{{Denom: BridgeDenom, Amount: math.OneInt()}}
	govGenState.Params.ExpeditedMinDeposit = sdk.Coins{{Denom: BridgeDenom, Amount: math.OneInt()}}
	bz, err := cdc.MarshalJSON(&govGenState)
	s.Require().NoError(err)
	appGenState[govtypes.ModuleName] = bz

	// set staking bond denom
	var stakingGenState stakingtypes.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[stakingtypes.ModuleName], &stakingGenState))
	bz, err = cdc.MarshalJSON(&stakingGenState)
	s.Require().NoError(err)
	appGenState[stakingtypes.ModuleName] = bz

	// set mint denom
	var mintGenState minttypes.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[minttypes.ModuleName], &mintGenState))
	mintGenState.Params.InflationMax = math.LegacyZeroDec()
	mintGenState.Params.InflationMin = math.LegacyZeroDec()
	mintGenState.Params.InflationRateChange = math.LegacyZeroDec()
	mintGenState.Minter.Inflation = math.LegacyZeroDec()
	bz, err = cdc.MarshalJSON(&mintGenState)
	s.Require().NoError(err)
	appGenState[minttypes.ModuleName] = bz

	// TODO: genesis supply will be incorrect if we add more accounts
	var bankGenState banktypes.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[banktypes.ModuleName], &bankGenState))
	genesisSupply := int64(len(s.chain.validators) * initBalance)
	bankGenState.Supply = sdk.NewCoins(sdk.NewCoin(BridgeDenom, math.NewInt(genesisSupply)))
	bz, err = cdc.MarshalJSON(&bankGenState)
	s.Require().NoError(err)
	appGenState[banktypes.ModuleName] = bz

	var bridgeGenState bridgetypes.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[bridgetypes.ModuleName], &bridgeGenState))
	bridgeGenState.Params.BridgeDenom = BridgeDenom
	bridgeGenState.Params.SupplyDeltaPeriod = SupplyDeltaPeriod
	bz, err = cdc.MarshalJSON(&bridgeGenState)
	s.Require().NoError(err)
	appGenState[bridgetypes.ModuleName] = bz

	var genUtilGenState genutiltypes.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[genutiltypes.ModuleName], &genUtilGenState))

	// generate genesis txs
	genTxs := make([]json.RawMessage, len(s.chain.validators))
	for i, val := range s.chain.validators {
		createValmsg, err := val.buildCreateValidatorMsg(InitStakedCoin)
		s.Require().NoError(err)

		signedTx, err := val.signMsg(createValmsg)
		s.Require().NoError(err)

		txRaw, err := cdc.MarshalJSON(signedTx)
		s.Require().NoError(err)

		genTxs[i] = txRaw
	}

	genUtilGenState.GenTxs = genTxs

	bz, err = cdc.MarshalJSON(&genUtilGenState)
	s.Require().NoError(err)
	appGenState[genutiltypes.ModuleName] = bz

	// serialize genesis state
	bz, err = json.MarshalIndent(appGenState, "", "  ")
	s.Require().NoError(err)

	genDoc.AppState = bz

	bz, err = cmjson.MarshalIndent(genDoc, "", "  ")
	s.Require().NoError(err)

	// write the updated genesis file to each validator
	for _, val := range s.chain.validators {
		s.Require().NoError(writeFile(filepath.Join(val.configDir(), "config", "genesis.json"), bz))
	}
}

func (s *E2ETestSuite) initFuelSequencerValidatorConfigs() {
	for i, val := range s.chain.validators {
		cmCfgPath := filepath.Join(val.configDir(), "config", "config.toml")

		vpr := viper.New()
		vpr.SetConfigFile(cmCfgPath)
		s.Require().NoError(vpr.ReadInConfig())

		valConfig := &cmconfig.Config{}
		s.Require().NoError(vpr.Unmarshal(valConfig))

		valConfig.P2P.ListenAddress = "tcp://0.0.0.0:26656"
		valConfig.P2P.AddrBookStrict = false
		valConfig.P2P.ExternalAddress = fmt.Sprintf("%s:%d", val.instanceName(), 26656)
		valConfig.RPC.ListenAddress = "tcp://0.0.0.0:26657"
		valConfig.StateSync.Enable = false
		valConfig.LogLevel = "info"

		// speed up blocks
		valConfig.Consensus.TimeoutCommit = 1 * time.Second
		valConfig.Consensus.TimeoutPropose = 1 * time.Second

		var peers []string

		for j := 0; j < len(s.chain.validators); j++ {
			if i == j {
				continue
			}

			peer := s.chain.validators[j]
			peerID := fmt.Sprintf("%s@%s%d:26656", peer.nodeKey.ID(), peer.moniker, j)
			peers = append(peers, peerID)
		}

		valConfig.P2P.PersistentPeers = strings.Join(peers, ",")

		cmconfig.WriteConfigFile(cmCfgPath, valConfig)

		// set application configuration
		appCfgPath := filepath.Join(val.configDir(), "config", "app.toml")

		appConfig := srvconfig.DefaultConfig()
		appConfig.API.Enable = true
		appConfig.API.Address = "tcp://0.0.0.0:1317"
		appConfig.GRPC.Address = "0.0.0.0:9090"
		appConfig.Pruning = "nothing"
		appConfig.MinGasPrices = fmt.Sprintf("%s%s", minGasPrices, BridgeDenom)

		srvconfig.WriteConfigFile(appCfgPath, appConfig)
	}
}

func (s *E2ETestSuite) runEthContainer() {
	s.T().Log("starting Ethereum container...")
	var err error
	runOpts := dockertest.RunOptions{
		Name:       "ethereum",
		Repository: ethereumDockerImageRepo,
		Tag:        ethereumDockerImageTag,
		NetworkID:  s.dockerNetwork.Network.ID,
		PortBindings: map[docker.Port][]docker.PortBinding{
			"8545/tcp": {{HostIP: "", HostPort: "8545"}},
		},
		ExposedPorts: []string{"8545/tcp"},
	}

	s.ethResource, err = s.dockerPool.RunWithOptions(
		&runOpts,
		noRestart,
	)
	s.Require().NoError(err)

	ethClient, err := ethclient.Dial(fmt.Sprintf("http://%s", s.ethResource.GetHostPort("8545/tcp")))
	s.Require().NoError(err)

	// Wait for the Ethereum node to respond to a request
	s.Require().Eventually(
		func() bool {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			balance, err := ethClient.BalanceAt(ctx, common.HexToAddress(s.chain.validators[0].ethereumKey.address), nil)
			if err != nil {
				s.T().Logf("error querying balance: %e", err)
				return false
			}

			if balance == nil {
				s.T().Logf("balance for first validator is nil")
			}

			if balance.Cmp(big.NewInt(0)) == 0 {
				s.T().Logf("balance for first validator is %s", balance.String())
				return false
			}

			return true
		},
		5*time.Minute,
		10*time.Second,
		"ethereum node failed to respond",
	)

	s.T().Logf("started Ethereum container: %s", s.ethResource.Container.ID)
}

func (s *E2ETestSuite) runFuelSequencerValidators() {
	s.T().Log("starting validator containers...")

	// Get user from OS to ensure permissions match up when the container writes files.
	user, err := osuser.Current()
	s.Require().NoError(err)

	s.valResources = make([]*dockertest.Resource, len(s.chain.validators))
	for i, val := range s.chain.validators {
		runOpts := &dockertest.RunOptions{
			Name:       val.instanceName(),
			NetworkID:  s.dockerNetwork.Network.ID,
			Repository: fuelSequencerDockerImageRepo,
			Tag:        fuelSequencerDockerImageTag,
			Mounts: []string{
				fmt.Sprintf("%s/:%s", val.configDir(), fuelSequencerValidatorDefaultHome),
			},
			User: fmt.Sprintf("%s:%s", user.Uid, user.Gid),
			// Assumption: image entrypoint is a script that runs a node and a sidecar.
		}

		// expose the first validator for debugging and communication
		if val.index == 0 {
			runOpts.PortBindings = map[docker.Port][]docker.PortBinding{
				"1317/tcp":  {{HostIP: "", HostPort: "1317"}},
				"9090/tcp":  {{HostIP: "", HostPort: "9090"}},
				"26656/tcp": {{HostIP: "", HostPort: "26656"}},
				"26657/tcp": {{HostIP: "", HostPort: "26657"}},
				"8080/tcp":  {{HostIP: "", HostPort: "8080"}},
			}
			runOpts.ExposedPorts = []string{"1317/tcp", "9090/tcp", "26656/tcp", "26657/tcp", "8080/tcp"}

			val.hostRPCPort = "tcp://localhost:26657"
			val.hostAPIPort = "tcp://localhost:1317"
			val.hostGRPCPort = "localhost:9090"
			val.sidecarGRPCPort = "localhost:8080"
		}

		resource, err := s.dockerPool.RunWithOptions(runOpts, noRestart)
		s.Require().NoError(err)

		s.valResources[i] = resource
		s.T().Logf("started validator container: %s", resource.Container.ID)
	}

	rpcClient, err := rpchttp.New("tcp://localhost:26657", "/websocket")
	s.Require().NoError(err)

	s.Require().Eventually(
		func() bool {
			status, err := rpcClient.Status(context.Background())
			if err != nil {
				s.T().Logf("can't get container status: %s", err.Error())
			}
			if status == nil {
				container, ok := s.dockerPool.ContainerByName("fuelsequencer0")
				if !ok {
					s.T().Logf("no container by 'fuelsequencer0'")
				} else {
					if container.Container.State.Status == "exited" {
						s.Fail("validators exited", "state: %s logs: \n%s", container.Container.State.String(), s.logsByContainerID(container.Container.ID))
						s.T().FailNow()
					}
					s.T().Logf("state: %v, health: %v", container.Container.State.Status, container.Container.State.Health)
				}
				return false
			}

			// let the node produce a few blocks
			if status.SyncInfo.CatchingUp {
				s.T().Logf("catching up: %t", status.SyncInfo.CatchingUp)
				return false
			}
			if status.SyncInfo.LatestBlockHeight < 2 {
				s.T().Logf("block height %d", status.SyncInfo.LatestBlockHeight)
				return false
			}

			return true
		},
		10*time.Minute,
		15*time.Second,
		"validator node failed to produce blocks",
	)
}

func noRestart(config *docker.HostConfig) {
	// in this case we don't want the nodes to restart on failure
	config.RestartPolicy = docker.RestartPolicy{
		Name: "no",
	}
}

func (s *E2ETestSuite) logsByContainerID(id string) string {
	var containerLogsBuf bytes.Buffer
	s.Require().NoError(s.dockerPool.Client.Logs(
		docker.LogsOptions{
			Container:    id,
			OutputStream: &containerLogsBuf,
			ErrorStream:  &containerLogsBuf,
			Stdout:       true,
			Stderr:       true,
		},
	))

	return containerLogsBuf.String()
}

func (s *E2ETestSuite) Ctx() context.Context {
	return context.Background()
}

func (s *E2ETestSuite) Logger() *zap.Logger {
	return s.log.With(
		zap.String("test", s.T().Name()),
	)
}
