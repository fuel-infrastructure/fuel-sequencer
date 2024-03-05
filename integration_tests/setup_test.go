package integration_tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
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
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fuel-infrastructure/fuel-sequencer/app"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

// TODO: teardown with shutting off of containers

func init() {
	app.InitSDKConfig()
	app.InitCometBFTConfig()
	app.InitAppConfig()

	sdk.DefaultBondDenom = testDenom
}

const (
	testDenom      = "ufuel"
	initBalanceStr = "210000000000ufuel"
	minGasPrice    = "2"

	fuelSequencerDockerImageRepo = "fuel-infrastructure/fuel-sequencer"
	fuelSequencerDockerImageTag  = "latest"

	ethereumDockerImageRepo = "fuel-infrastructure/contracts-docker-e2e"
	ethereumDockerImageTag  = "latest"
)

var (
	stakeAmount, _  = math.NewIntFromString("100000000000")
	stakeAmountCoin = sdk.NewCoin(testDenom, stakeAmount)
)

func MNEMONICS() []string {
	return []string{
		"test test test test test test test test test test test junk", // TODO: must match one on test-contracts
		"receive roof marine sure lady hundred sea enact exist place bean wagon kingdom betray science photo loop funny bargain floor suspect only strike endless",
		"march carpet enact kiss tribe plastic wash enter index lift topic riot try juice replace supreme original shift hover adapt mutual holiday manual nut",
		"assault section bleak gadget venture ship oblige pave fabric more initial april dutch scene parade shallow educate gesture lunar match patch hawk member problem",
		"say monitor orient heart super local purse cricket caution primary bring insane road expect rather help two extend own execute throw nation plunge subject",
	}
}

type IntegrationTestSuite struct {
	suite.Suite

	chain         *chain
	dockerPool    *dockertest.Pool
	dockerNetwork *dockertest.Network
	ethResource   *dockertest.Resource
	valResources  []*dockertest.Resource
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up e2e integration test suite...")

	var err error
	s.chain, err = newChain()
	s.Require().NoError(err)
	s.dockerPool, err = dockertest.NewPool("")
	s.Require().NoError(err)
	s.dockerNetwork, err = s.dockerPool.CreateNetwork(fmt.Sprintf("%s-testnet", s.chain.id))
	s.Require().NoError(err)

	s.T().Logf("starting e2e infrastructure; chain-id: %s; datadir: %s", s.chain.id, s.chain.dataDir)

	// initialization
	mnemonics := MNEMONICS()
	s.initNodesWithMnemonics(mnemonics...)
	s.initEthereumFromMnemonics(mnemonics)

	// run the eth container so that the contract addresses are available
	//s.runEthContainer()

	// continue generating node genesis
	s.initGenesis()
	s.initValidatorConfigs()

	// container infrastructure
	s.runValidators()
}

func (s *IntegrationTestSuite) TearDownSuite() {
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

func (s *IntegrationTestSuite) initNodes(nodeCount int) { //nolint:unused
	s.Require().NoError(s.chain.createAndInitValidators(nodeCount))

	// initialize a genesis file for the first validator
	val0ConfigDir := s.chain.validators[0].configDir()
	for _, val := range s.chain.validators {
		s.Require().NoError(
			addGenesisAccount(val0ConfigDir, "", initBalanceStr, val.address()),
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

func (s *IntegrationTestSuite) initNodesWithMnemonics(mnemonics ...string) {
	s.Require().NoError(s.chain.createAndInitValidatorsWithMnemonics(mnemonics))

	//initialize a genesis file for the first validator
	val0ConfigDir := s.chain.validators[0].configDir()
	for _, val := range s.chain.validators {
		// Fund the first validator with some funds to be used by auction module integration tests
		s.Require().NoError(
			addGenesisAccount(val0ConfigDir, "", initBalanceStr, val.address()),
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

func (s *IntegrationTestSuite) initEthereum() { //nolint:unused
	// TODO: create genesis file instead of assuming it exists

	for _, val := range s.chain.validators {
		s.Require().NoError(val.generateEthereumKey())
	}
}

func (s *IntegrationTestSuite) initEthereumFromMnemonics(mnemonics []string) {
	// TODO: create genesis file instead of assuming it exists

	for i, val := range s.chain.validators {
		s.Require().NoError(val.generateEthereumKeyFromMnemonic(mnemonics[i]))
	}
}

func (s *IntegrationTestSuite) initGenesis() {
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
	govGenState.Params.MinDeposit = sdk.Coins{{Denom: testDenom, Amount: math.OneInt()}}
	govGenState.Params.ExpeditedMinDeposit = sdk.Coins{{Denom: testDenom, Amount: math.OneInt()}}
	bz, err := cdc.MarshalJSON(&govGenState)
	s.Require().NoError(err)
	appGenState[govtypes.ModuleName] = bz

	// set crisis denom
	var crisisGenState crisistypes.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[crisistypes.ModuleName], &crisisGenState))
	crisisGenState.ConstantFee.Denom = testDenom
	bz, err = cdc.MarshalJSON(&crisisGenState)
	s.Require().NoError(err)
	appGenState[crisistypes.ModuleName] = bz

	// set staking bond denom
	var stakingGenState stakingtypes.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[stakingtypes.ModuleName], &stakingGenState))
	stakingGenState.Params.BondDenom = testDenom
	bz, err = cdc.MarshalJSON(&stakingGenState)
	s.Require().NoError(err)
	appGenState[stakingtypes.ModuleName] = bz

	// set mint denom
	var mintGenState minttypes.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[minttypes.ModuleName], &mintGenState))
	mintGenState.Params.MintDenom = testDenom
	mintGenState.Params.InflationMax = math.LegacyZeroDec()
	mintGenState.Params.InflationMin = math.LegacyZeroDec()
	mintGenState.Params.InflationRateChange = math.LegacyZeroDec()
	mintGenState.Minter.Inflation = math.LegacyZeroDec()
	bz, err = cdc.MarshalJSON(&mintGenState)
	s.Require().NoError(err)
	appGenState[minttypes.ModuleName] = bz

	//distGenState := disttypes.DefaultGenesisState()
	//distGenState.Params.CommunityTax = math.LegacyZeroDec()
	//distGenState.FeePool.CommunityPool = sdk.NewDecCoins(sdk.NewDecCoin(testDenom, math.NewInt(1000000000)))
	//bz, err = cdc.MarshalJSON(distGenState)
	//s.Require().NoError(err)
	//appGenState[disttypes.ModuleName] = bz

	var bankGenState banktypes.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[banktypes.ModuleName], &bankGenState))
	bankGenState.Supply = sdk.NewCoins(sdk.NewCoin(testDenom, math.NewInt(101050001000000))) // TODO: make dynamic
	bz, err = cdc.MarshalJSON(&bankGenState)
	s.Require().NoError(err)
	appGenState[banktypes.ModuleName] = bz

	var genUtilGenState genutiltypes.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[genutiltypes.ModuleName], &genUtilGenState))

	// TODO: bridge module state
	// TODO: sequencing module state

	// generate genesis txs
	genTxs := make([]json.RawMessage, len(s.chain.validators))
	for i, val := range s.chain.validators {
		createValmsg, err := val.buildCreateValidatorMsg(stakeAmountCoin)
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

func (s *IntegrationTestSuite) initValidatorConfigs() {
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
		appConfig.Pruning = "nothing"
		appConfig.MinGasPrices = fmt.Sprintf("%s%s", minGasPrice, testDenom)

		srvconfig.WriteConfigFile(appCfgPath, appConfig)
	}
}

func (s *IntegrationTestSuite) runEthContainer() {
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

func (s *IntegrationTestSuite) runValidators() {
	s.T().Log("starting validator containers...")

	s.valResources = make([]*dockertest.Resource, len(s.chain.validators))
	for i, val := range s.chain.validators {
		runOpts := &dockertest.RunOptions{
			Name:       val.instanceName(),
			NetworkID:  s.dockerNetwork.Network.ID,
			Repository: fuelSequencerDockerImageRepo,
			Tag:        fuelSequencerDockerImageTag,
			Mounts: []string{
				fmt.Sprintf("%s/:/home/fuelsequencer/.fuelsequencerd", val.configDir()),
			},
			// TODO: why do we have to specify home explicitly? what's the default home?
			Entrypoint: []string{"fuelsequencerd", "start", "--trace=true", "--home", "/home/fuelsequencer/.fuelsequencerd"},
		}

		// expose the first validator for debugging and communication
		if val.index == 0 {
			runOpts.PortBindings = map[docker.Port][]docker.PortBinding{
				"1317/tcp":  {{HostIP: "", HostPort: "1317"}},
				"9090/tcp":  {{HostIP: "", HostPort: "9090"}},
				"26656/tcp": {{HostIP: "", HostPort: "26656"}},
				"26657/tcp": {{HostIP: "", HostPort: "26657"}},
			}
			runOpts.ExposedPorts = []string{"1317/tcp", "9090/tcp", "26656/tcp", "26657/tcp"}
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

func (s *IntegrationTestSuite) logsByContainerID(id string) string {
	var containerLogsBuf bytes.Buffer
	s.Require().NoError(s.dockerPool.Client.Logs(
		docker.LogsOptions{
			Container:    id,
			OutputStream: &containerLogsBuf,
			Stdout:       true,
			Stderr:       true,
		},
	))

	return containerLogsBuf.String()
}

func (s *IntegrationTestSuite) TestBasicChain() {
	// this test verifies that the setup functions all operate as expected
	s.Run("bring up basic chain", func() {
	})
}
