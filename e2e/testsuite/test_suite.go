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
	BridgeDenom = "utest"

	// Gas configs
	minGasPrices = "0.01"
	defaultTxGas = 1000000

	// Genesis configs
	supplyDeltaPeriod      = uint64(10)       // default - can be overridden
	governanceVotingPeriod = time.Second * 20 // default - can be overridden

	// Balance and staked amount per validator
	initBalance = 210000000000 // per validator
	initStaked  = 100000000000 // per validator

	// FuelSequencer validator configs
	fuelSequencerValidatorDefaultHome = "/home/fuelsequencer/.fuelsequencer"
	fuelSequencerBinary               = "fuelsequencerd"

	// Docker configs
	fuelSequencerDockerImageRepo      = "fuel-infrastructure/fuel-sequencer"
	fuelSequencerDockerImageTag       = "latest"
	ethereumNodeDockerImageRepo       = "ghcr.io/foundry-rs/foundry"
	ethereumNodeDockerImageTag        = "nightly"
	ethereumDeploymentDockerImageRepo = "fuel-rollup/ethereum-deployment"
	ethereumDeploymentDockerImageTag  = "latest"

	// TestSuite configs
	blocksToWaitForGovProposalToPass = uint64(25)

	// GuardianPrivateKey is the private key of the address assigned as the 'guardian'.
	GuardianPrivateKey = "0xdf57089febbacf7ba0bc227dafbffa9fc08a93fdc68e1e42411a14efcf23656e"
	// FuelStreamXContractAddress is the address of the contract that holds bridge commitments.
	FuelStreamXContractAddress = "0x959922bE3CAee4b8Cd9a407cc3ac1C251C2007B1"
	// TokenContractAddress is the address of the FUEL token contract.
	TokenContractAddress = "0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512"
	// SequencerInterfaceContractAddress is the address of the contract that has the batchAuthorize function.
	SequencerInterfaceContractAddress = "0x2279B7A0a67DB372996a5FaB50D91eAA73d2eBe6"
)

var (
	// Balance and staked amount per validator
	InitBalanceCoin = sdk.NewInt64Coin(BridgeDenom, initBalance)
	InitStakedCoin  = sdk.NewInt64Coin(BridgeDenom, initStaked)

	// MNEMONICS dictates how many Sequencer nodes will be created by specifying their mnemonic.
	// The first mnemonic is reused for the Ethereum validator mnemonic.
	MNEMONICS = []string{
		// This corresponds to the typically used signer in the contracts.
		"test test test test test test test test test test test junk",
		// Alice
		"dinner crash nurse casino baby fold race cheese elite column sausage sleep close royal rain over mechanic minimum outdoor conduct cash wagon frog evidence",
		// Bob
		"gaze drama excess raven follow antenna swallow beef upper myself question pitch course ill adult century crisp ice rough match praise sing unveil vintage",
	}

	// Inflation params
	InflationRateChange = sdkmath.LegacyMustNewDecFromStr("0.13")
	InflationMax        = sdkmath.LegacyMustNewDecFromStr("0.2")
	InflationMin        = sdkmath.LegacyMustNewDecFromStr("0.07")

	// Vesting params
	VestingStartTimeDelay = time.Hour * 24 * 365

	// Logging
	LogLevel = zaptest.Level(zap.DebugLevel)
)

type E2ETestSuite struct {
	suite.Suite

	log *zap.Logger

	Chain         *chain
	dockerPool    *dockertest.Pool
	dockerNetwork *dockertest.Network

	ethNodeResource       *dockertest.Resource
	ethDeploymentResource *dockertest.Resource
	valResources          []*dockertest.Resource

	// govProposalIdCounter keeps track of the latest governance proposal ID, so we can vote using the ID.
	govProposalIdCounter int

	// genesisOverrides are extra changes applied to genesis before the Sequencer is started. For these overrides to be
	// applied, SetupTest needs to be overridden so that genesisOverrides can be changed before invoking SetupTest.
	GenesisOverrides *ModifyGenesisFunc

	// SeqKeys are the FuelSequencer wallets derived from the above MNEMONICS, with hex versions of the addresses.
	// This is filled-in later on in SetupTest, once the address codec has been initialised.
	SeqKeys []*SequencerKey

	// EthKeys are the Ethereum wallets derived from the above MNEMONICS, with Bech32 versions of the addresses.
	// This is filled-in later on in SetupTest, once the address codec has been initialised.
	EthKeys []*EthereumKey

	// EthGuardian corresponds to the address assigned as the guardian in the FuelStreamX contract.
	// Ref: https://github.com/FuelLabs/fuel-rollup/blob/main/deploy/hardhat/011.set_fuelstreamx_config.ts#L42-L48
	// This is the private key at index 19 in the list of private keys generated by anvil, available in the anvil logs.
	// This is filled-in later on in SetupTest, once the address codec has been initialised.
	EthGuardian *EthereumKey
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

	// Derive and print the Ethereum keys with the hex and bech32 representation of the addresses.
	for _, mnemonic := range MNEMONICS {
		s.EthKeys = append(s.EthKeys, mustNewEthereumKeyFromMnemonic(mnemonic))
	}
	s.EthGuardian = mustNewEthereumKeyFromPrivateKey(GuardianPrivateKey)
	s.T().Logf("ethereum keys:")
	for _, key := range append(s.EthKeys, s.EthGuardian) {
		s.T().Logf("\tpriv:%s hex:%s seq:%s", key.PrivateKeyHex, key.AddressHex, key.AddressSeq)
	}

	// Derive and print the Sequencer keys with the hex and bech32 representation of the addresses.
	for _, mnemonic := range MNEMONICS {
		s.SeqKeys = append(s.SeqKeys, mustNewSequencerKeyFromMnemonic(mnemonic))
	}
	s.T().Logf("sequencer keys:")
	for _, key := range s.SeqKeys {
		s.T().Logf("\tacc:%s val:%s hex:%s", key.AddressSeq, key.ValAddressSeq, key.AddressHex)
	}

	// initialization
	s.initFuelSequencerNodes(MNEMONICS)

	// run Ethereum node
	s.runEthereumNodeContainer()

	// run FuelSequencer nodes and sidecars
	s.initFuelSequencerGenesis()
	s.initFuelSequencerValidatorConfigs()
	s.runFuelSequencerValidators()

	// deploy Ethereum contracts
	s.runEthereumDeploymentContainer()

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

	// Set the genesis header
	data := PackUpdateGenesisStateMessage(1, common.BytesToHash(genesisBlockHeaderHash))
	_, err = s.SendEthTransactionToFuelStreamXContract(data)
	s.Require().NoError(err)

	// Ensure that FuelSequencer has approximately caught up with Ethereum.
	// This assumes that the FuelSequencer has a shorter block time than Ethereum.
	for {
		fromHeight := s.QueryLastEthereumBlockSynced(s.Ctx())
		toHeight, err := s.getEthereumRPCClient().BlockNumber(s.Ctx())
		s.Require().NoError(err)
		if fromHeight+5 > toHeight { // max 5 blocks difference
			break
		}

		s.T().Logf("waiting for FuelSequencer to sync to Ethereum (%d -> %d)...", fromHeight, toHeight)
		s.PollForLastEthereumBlockSynced(s.Ctx(), 50, toHeight)
	}

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
	s.Require().NoError(s.dockerPool.Purge(s.ethNodeResource))
	s.Require().NoError(s.dockerPool.Purge(s.ethDeploymentResource))

	for _, vc := range s.valResources {
		s.Require().NoError(s.dockerPool.Purge(vc))
	}

	s.Require().NoError(s.dockerPool.RemoveNetwork(s.dockerNetwork))

	s.govProposalIdCounter = 1
	s.GenesisOverrides = nil

	s.SeqKeys = nil
	s.EthKeys = nil
	s.EthGuardian = nil
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

func (s *E2ETestSuite) runEthereumNodeContainer() {
	s.T().Log("starting Ethereum node container...")
	var err error
	runOpts := dockertest.RunOptions{
		Name:       "ethereum-node",
		Repository: ethereumNodeDockerImageRepo,
		Tag:        ethereumNodeDockerImageTag,
		NetworkID:  s.dockerNetwork.Network.ID,
		PortBindings: map[docker.Port][]docker.PortBinding{
			"8545/tcp": {{HostIP: "", HostPort: "8545"}},
		},
		ExposedPorts: []string{"8545/tcp"},
		Entrypoint: []string{
			"anvil",
			"--host", "0.0.0.0",
			"--mnemonic", MNEMONICS[0],
			"--accounts", "20",
			"--slots-in-an-epoch", "1",
			// Note: do not set --block-time since this is overridden by the deployment scripts.
		},
	}

	s.ethNodeResource, err = s.dockerPool.RunWithOptions(
		&runOpts,
		noRestart,
	)
	s.Require().NoError(err)

	ethClient, err := ethclient.Dial(fmt.Sprintf("http://%s", s.ethNodeResource.GetHostPort("8545/tcp")))
	s.Require().NoError(err)

	// Wait for the Ethereum node to respond to a request
	s.Require().Eventually(
		func() bool {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			balance, err := ethClient.BalanceAt(ctx, s.EthKeys[0].Address, nil)
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

	s.T().Logf("started Ethereum node container: %s", s.ethNodeResource.Container.ID)
}

func (s *E2ETestSuite) runEthereumDeploymentContainer() {
	s.T().Log("starting Ethereum deployment container to deploy contracts...")
	var err error
	runOpts := dockertest.RunOptions{
		Name:       "ethereum-deployment",
		Repository: ethereumDeploymentDockerImageRepo,
		Tag:        ethereumDeploymentDockerImageTag,
		NetworkID:  s.dockerNetwork.Network.ID,
		Env: []string{
			"RPC_URL=http://ethereum-node:8545",
		},
		Cmd: []string{"npx", "hardhat", "deploy", "--network", "localhost", "--reset"},
	}

	s.ethDeploymentResource, err = s.dockerPool.RunWithOptions(
		&runOpts,
		noRestart,
	)
	s.Require().NoError(err)

	// Wait for the contracts to be deployed, i.e. for the deployment container to stop
	s.T().Logf("waiting for Ethereum contracts to be deployed...")
	waitContext, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	_, err = s.dockerPool.Client.WaitContainerWithContext(s.ethDeploymentResource.Container.ID, waitContext)
	s.Require().NoError(err)
}

func (s *E2ETestSuite) PauseEthereum() {
	s.Require().NoError(s.dockerPool.Client.PauseContainer(s.ethNodeResource.Container.ID))
}

func (s *E2ETestSuite) UnpauseEthereum() {
	s.Require().NoError(s.dockerPool.Client.UnpauseContainer(s.ethNodeResource.Container.ID))
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
