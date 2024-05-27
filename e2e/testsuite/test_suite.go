// Test suite inspired by https://github.com/PeggyJV/sommelier/tree/v7.0.1/integration_tests

package testsuite

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"os"
	osuser "os/user"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	cmtbytes "github.com/cometbft/cometbft/libs/bytes"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fuel-infrastructure/fuel-sequencer/app"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
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
	defaultTxGas      = 1000000
	supplyDeltaPeriod = uint64(10) // default - can be overridden

	// Balance and staked amount per validator
	initBalance = 210000000000 // per validator
	initStaked  = 100000000000 // per validator

	fuelSequencerDockerImageRepo = "fuel-infrastructure/fuel-sequencer"
	fuelSequencerDockerImageTag  = "latest"

	fuelSequencerValidatorDefaultHome = "/home/fuelsequencer/.fuelsequencer"
	fuelSequencerBinary               = "fuelsequencerd"

	ethereumDockerImageRepo = "fuel-infrastructure/contracts-docker-e2e"
	ethereumDockerImageTag  = "latest"

	governanceVotingPeriod           = time.Second * 20 // default - can be overridden
	blocksToWaitForGovProposalToPass = uint64(25)

	fuelStreamXDockerImageRepo = "fuel-infrastructure/fuel-stream-x-manual-docker-e2e"
	fuelStreamXDockerImageTag  = "latest"
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

	// ETH_ADDRESSES are the Ethereum addresses derived from the above MNEMONICS.
	ETH_ADDRESSES = []string{
		"0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
		"0xe53e6e952cf156b9f58a2a82da5ea537102ba484",
		"0x8fe6350f77cf9be08bbac2c8156caba4d47e756b",
	}

	// ETH_ADDRESS_SEQ are the addresses mapped from ETH_ADDRESSES on the Sequencer
	ETH_ADDRESS_SEQ = []string{
		"fuelsequencer17w0adeg64ky0daxwd2ugyuneellmjgnx5dpmtz",
		"fuelsequencer1u5lxa9fv79ttnav292pd5h49xugzhfyyzgs97h",
		"fuelsequencer13lnr2rmhe7d7pza6ctyp2m9t5n28uattwk7kr8",
	}

	// FUEL_STREAM_X_CONTRACT is the FuelStreamX contract that generates events, deployed on the Ethereum node.
	FUEL_STREAM_X_CONTRACT = "0xa513E6E4b8f2a923D98304ec87F64353C4D5C853"
	// UPDATE_DELAY_BLOCKS is the block interval at which FuelStreamX submits bridge commitments to Ethereum.
	UPDATE_DELAY_BLOCKS = 25

	// Inflation params
	InflationRateChange = sdkmath.LegacyMustNewDecFromStr("0.13")
	InflationMax        = sdkmath.LegacyMustNewDecFromStr("0.2")
	InflationMin        = sdkmath.LegacyMustNewDecFromStr("0.07")

	// Vesting params
	VestingStartTimeDelay = time.Hour * 24 * 365
)

var (
	LogLevel = zaptest.Level(zap.DebugLevel)
)

type E2ETestSuite struct {
	suite.Suite

	log *zap.Logger

	Chain         *chain
	dockerPool    *dockertest.Pool
	dockerNetwork *dockertest.Network

	ethResource         *dockertest.Resource
	valResources        []*dockertest.Resource
	fuelStreamXResource *dockertest.Resource

	// govProposalIdCounter keeps track of the latest governance proposal ID, so we can vote using the ID.
	govProposalIdCounter int

	// genesisOverrides are extra changes applied to genesis before the Sequencer is started. For these overrides to be
	// applied, SetupTest needs to be overridden so that genesisOverrides can be changed before invoking SetupTest.
	GenesisOverrides *ModifyGenesisFunc
}

func (s *E2ETestSuite) SetupTest() {
	if testing.Short() {
		s.T().Skip()
	}

	s.T().Log("setting up E2E test...")

	s.log = zaptest.NewLogger(s.T(), LogLevel)

	var err error
	s.Chain, err = newChain(len(MNEMONICS))
	s.Require().NoError(err)
	s.dockerPool, err = dockertest.NewPool("")
	s.Require().NoError(err)
	s.dockerNetwork, err = s.dockerPool.CreateNetwork(fmt.Sprintf("%s-testnet", s.Chain.id))
	s.Require().NoError(err)

	s.T().Logf("starting E2E infrastructure; Chain-id: %s; datadir: %s", s.Chain.id, s.Chain.dataDir)

	// initialization
	s.initFuelSequencerNodes(MNEMONICS)
	s.initEthereumNodes(MNEMONICS)

	// Run the eth container, no contracts deployed yet just anvil
	s.runEthContainer()

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
	err = s.WaitForSequencerBlocks(s.Ctx(), 1, time.Minute)
	s.Require().NoError(err)

	// Get genesis header
	genesisBlockHeaderHash, err := s.Chain.GetBlockHeaderHash(s.Ctx(), 1)
	s.Require().NoError(err)

	// Deploy the contracts with the header
	s.deployContracts(1, genesisBlockHeaderHash)

	// Reset the proposal counter since we're starting a new chain.
	s.govProposalIdCounter = 1
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

	s.Require().NoError(os.RemoveAll(s.Chain.dataDir))
	s.Require().NoError(s.dockerPool.Purge(s.ethResource))

	for _, vc := range s.valResources {
		s.Require().NoError(s.dockerPool.Purge(vc))
	}

	// Operator and relayer should have been purged earlier, but purge just in case
	if s.fuelStreamXResource != nil {
		_ = s.dockerPool.Purge(s.fuelStreamXResource)
	}

	s.Require().NoError(s.dockerPool.RemoveNetwork(s.dockerNetwork))

	s.govProposalIdCounter = 1
	s.GenesisOverrides = nil
}

// initFuelSequencerNodes initialises FuelSequencer nodes with mnemonics (if specified) or random keys.
// It also sets up the genesis file using the first validator and copies it to all other validator nodes.
func (s *E2ETestSuite) initFuelSequencerNodes(mnemonics []string) {
	s.Require().NoError(s.Chain.createAndInitFuelSequencerValidators(mnemonics))

	// initialize a genesis file for the first validator
	val0ConfigDir := s.Chain.validators[0].configDir()
	for _, val := range s.Chain.validators {
		s.Require().NoError(
			addGenesisAccount(val0ConfigDir, "", InitBalanceCoin.String(), val.address()),
		)
	}

	// copy the genesis file to the remaining validators
	for _, val := range s.Chain.validators[1:] {
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

	for i, val := range s.Chain.validators {
		if useMnemonics {
			s.Require().NoError(val.generateEthereumKeyFromMnemonic(mnemonics[i]))
		} else {
			s.Require().NoError(val.generateEthereumKey())
		}
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

			balance, err := ethClient.BalanceAt(ctx, common.HexToAddress(s.Chain.validators[0].ethereumKey.address), nil)
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
		1*time.Minute,
		1*time.Second,
		"ethereum node failed to respond",
	)

	s.T().Logf("started Ethereum container: %s", s.ethResource.Container.ID)
}

func (s *E2ETestSuite) PauseEthereum() {
	s.Require().NoError(s.dockerPool.Client.PauseContainer(s.ethResource.Container.ID))
}

func (s *E2ETestSuite) UnpauseEthereum() {
	s.Require().NoError(s.dockerPool.Client.UnpauseContainer(s.ethResource.Container.ID))
}

func (s *E2ETestSuite) deployContracts(genesisHeight uint64, genesisHeaderHash cmtbytes.HexBytes) {
	s.T().Log("deploying Ethereum contracts...")

	execOptions := dockertest.ExecOptions{
		Env: []string{
			"RPC_URL=http://ethereum:8545",
			fmt.Sprintf("PRIVATE_KEY=%s", s.GetEthPrivateKeyHex()),
			"GUARDIAN_ADDRESS=0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266", // Can be anything
			fmt.Sprintf("GENESIS_HEIGHT=%d", genesisHeight),
			fmt.Sprintf("GENESIS_HEADER=%s", genesisHeaderHash.String()),
		},
	}

	exitCode, err := s.ethResource.Exec(
		[]string{"bash", "scripts/deploy_contract.sh"},
		execOptions,
	)
	s.Require().NoError(err)
	s.Require().Zero(exitCode)

	// Wait for the Ethereum node to respond to a request
	s.Require().Eventually(
		func() bool {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			code, err := s.Chain.ethClient.CodeAt(ctx, common.HexToAddress(FUEL_STREAM_X_CONTRACT), nil)
			if err != nil {
				s.T().Logf("error retreiving contract's code: %e", err)
				return false
			} else if len(code) == 0 {
				s.T().Logf("error retreiving contract's code, contract not depeloyed")
				return false
			}

			return true
		},
		1*time.Minute,
		1*time.Second,
		"ethereum node failed to respond",
	)

	s.T().Logf("deployed Ethereum contracts: %s", s.ethResource.Container.ID)
}

func (s *E2ETestSuite) runFuelSequencerValidators() {
	s.T().Log("starting validator containers...")

	// Get user from OS to ensure permissions match up when the container writes files.
	user, err := osuser.Current()
	s.Require().NoError(err)

	s.valResources = make([]*dockertest.Resource, len(s.Chain.validators))
	for i, val := range s.Chain.validators {
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
		}

		resource, err := s.dockerPool.RunWithOptions(runOpts, noRestart)
		s.Require().NoError(err)

		port1317 := resource.Container.NetworkSettings.Ports["1317/tcp"][0].HostPort
		val.hostAPIPort = fmt.Sprintf("tcp://localhost:%s", port1317)
		port9090 := resource.Container.NetworkSettings.Ports["9090/tcp"][0].HostPort
		val.hostGRPCPort = fmt.Sprintf("localhost:%s", port9090)
		port26657 := resource.Container.NetworkSettings.Ports["26657/tcp"][0].HostPort
		val.hostRPCPort = fmt.Sprintf("tcp://localhost:%s", port26657)
		port8080 := resource.Container.NetworkSettings.Ports["8080/tcp"][0].HostPort
		val.sidecarGRPCPort = fmt.Sprintf("localhost:%s", port8080)

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
		1*time.Minute,
		1*time.Second,
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
