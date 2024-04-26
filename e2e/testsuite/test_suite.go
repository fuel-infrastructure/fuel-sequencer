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
	// GATEWAY_CONTRACT is a contract by Succinct that does ZK proof verification.
	GATEWAY_CONTRACT = "0xDc64a140Aa3E981100a9becA4E685f962f0cF6C9"

	// Circuits
	NEXT_HEADER_FUNCTION_ID  = "0xbc40fbf4394cd00f78fae9763b0c2c71b21ea442c42fdadc5b720537240ebac1"
	HEADER_RANGE_FUNCTION_ID = "0xa3c1274aadd82e4d12c8004c33fb244ca686dad4fcc8957fc5668588c11d9502"

	// ABI
	FUEL_STREAM_X_ABI = `[{"type":"constructor","inputs":[{"name":"_params","type":"tuple","internalType":"structFuelStreamX.InitParameters","components":[{"name":"guardian","type":"address","internalType":"address"},{"name":"gateway","type":"address","internalType":"address"},{"name":"height","type":"uint64","internalType":"uint64"},{"name":"header","type":"bytes32","internalType":"bytes32"},{"name":"nextHeaderFunctionId","type":"bytes32","internalType":"bytes32"},{"name":"headerRangeFunctionId","type":"bytes32","internalType":"bytes32"}]}],"stateMutability":"nonpayable"},{"type":"function","name":"Authorize","inputs":[{"name":"_message","type":"bytes","internalType":"bytes"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"BRIDGE_COMMITMENT_MAX","inputs":[],"outputs":[{"name":"","type":"uint64","internalType":"uint64"}],"stateMutability":"view"},{"type":"function","name":"VERSION","inputs":[],"outputs":[{"name":"","type":"string","internalType":"string"}],"stateMutability":"pure"},{"type":"function","name":"blockHeightToHeaderHash","inputs":[{"name":"","type":"uint64","internalType":"uint64"}],"outputs":[{"name":"","type":"bytes32","internalType":"bytes32"}],"stateMutability":"view"},{"type":"function","name":"commitHeaderRange","inputs":[{"name":"_targetBlock","type":"uint64","internalType":"uint64"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"commitNextHeader","inputs":[{"name":"_trustedBlock","type":"uint64","internalType":"uint64"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"deposit","inputs":[{"name":"_amount","type":"uint256","internalType":"uint256"},{"name":"_to","type":"string","internalType":"string"},{"name":"_duration","type":"uint256","internalType":"uint256"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"frozen","inputs":[],"outputs":[{"name":"","type":"bool","internalType":"bool"}],"stateMutability":"view"},{"type":"function","name":"gateway","inputs":[],"outputs":[{"name":"","type":"address","internalType":"address"}],"stateMutability":"view"},{"type":"function","name":"headerRangeFunctionId","inputs":[],"outputs":[{"name":"","type":"bytes32","internalType":"bytes32"}],"stateMutability":"view"},{"type":"function","name":"latestBlock","inputs":[],"outputs":[{"name":"","type":"uint64","internalType":"uint64"}],"stateMutability":"view"},{"type":"function","name":"nextHeaderFunctionId","inputs":[],"outputs":[{"name":"","type":"bytes32","internalType":"bytes32"}],"stateMutability":"view"},{"type":"function","name":"processSequencerWithdrawalMessage","inputs":[{"name":"_proofNonce","type":"uint256","internalType":"uint256"},{"name":"bridgeCommitmentLeaf","type":"tuple","internalType":"structFuelStreamX.BridgeCommitmentLeaf","components":[{"name":"height","type":"uint256","internalType":"uint256"},{"name":"resultsHash","type":"bytes32","internalType":"bytes32"}]},{"name":"bridgeCommitmentLeafProof","type":"tuple","internalType":"structBinaryMerkleProof","components":[{"name":"sideNodes","type":"bytes32[]","internalType":"bytes32[]"},{"name":"key","type":"uint256","internalType":"uint256"},{"name":"numLeaves","type":"uint256","internalType":"uint256"}]},{"name":"txResultMarshalled","type":"bytes","internalType":"bytes"},{"name":"txResultProof","type":"tuple","internalType":"structBinaryMerkleProof","components":[{"name":"sideNodes","type":"bytes32[]","internalType":"bytes32[]"},{"name":"key","type":"uint256","internalType":"uint256"},{"name":"numLeaves","type":"uint256","internalType":"uint256"}]}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"requestHeaderRange","inputs":[{"name":"_targetBlock","type":"uint64","internalType":"uint64"}],"outputs":[],"stateMutability":"payable"},{"type":"function","name":"requestNextHeader","inputs":[],"outputs":[],"stateMutability":"payable"},{"type":"function","name":"setBridgeCommitmentRoot","inputs":[{"name":"nonce","type":"uint256","internalType":"uint256"},{"name":"bridgeCommitmentRoot","type":"bytes32","internalType":"bytes32"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"state_dataCommitments","inputs":[{"name":"","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"","type":"bytes32","internalType":"bytes32"}],"stateMutability":"view"},{"type":"function","name":"state_proofNonce","inputs":[],"outputs":[{"name":"","type":"uint256","internalType":"uint256"}],"stateMutability":"view"},{"type":"function","name":"updateFreeze","inputs":[{"name":"_freeze","type":"bool","internalType":"bool"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"updateFunctionIds","inputs":[{"name":"_headerRangeFunctionId","type":"bytes32","internalType":"bytes32"},{"name":"_nextHeaderFunctionId","type":"bytes32","internalType":"bytes32"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"updateGateway","inputs":[{"name":"_gateway","type":"address","internalType":"address"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"updateGenesisState","inputs":[{"name":"_height","type":"uint32","internalType":"uint32"},{"name":"_header","type":"bytes32","internalType":"bytes32"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"withdrawalNoncesExecuted","inputs":[{"name":"","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"","type":"bool","internalType":"bool"}],"stateMutability":"view"},{"type":"event","name":"AuthorizeEvent","inputs":[{"name":"_from","type":"address","indexed":true,"internalType":"address"},{"name":"_message","type":"bytes","indexed":false,"internalType":"bytes"}],"anonymous":false},{"type":"event","name":"DataCommitmentStored","inputs":[{"name":"proofNonce","type":"uint256","indexed":false,"internalType":"uint256"},{"name":"startBlock","type":"uint64","indexed":true,"internalType":"uint64"},{"name":"endBlock","type":"uint64","indexed":true,"internalType":"uint64"},{"name":"dataCommitment","type":"bytes32","indexed":true,"internalType":"bytes32"}],"anonymous":false},{"type":"event","name":"HeadUpdate","inputs":[{"name":"blockNumber","type":"uint64","indexed":false,"internalType":"uint64"},{"name":"headerHash","type":"bytes32","indexed":false,"internalType":"bytes32"}],"anonymous":false},{"type":"event","name":"HeaderRangeRequested","inputs":[{"name":"trustedBlock","type":"uint64","indexed":true,"internalType":"uint64"},{"name":"trustedHeader","type":"bytes32","indexed":true,"internalType":"bytes32"},{"name":"targetBlock","type":"uint64","indexed":true,"internalType":"uint64"}],"anonymous":false},{"type":"event","name":"NextHeaderRequested","inputs":[{"name":"trustedBlock","type":"uint64","indexed":true,"internalType":"uint64"},{"name":"trustedHeader","type":"bytes32","indexed":true,"internalType":"bytes32"}],"anonymous":false},{"type":"event","name":"SendToSequencerEvent","inputs":[{"name":"_from","type":"address","indexed":true,"internalType":"address"},{"name":"_amount","type":"uint256","indexed":false,"internalType":"uint256"},{"name":"_to","type":"string","indexed":false,"internalType":"string"},{"name":"_duration","type":"uint256","indexed":false,"internalType":"uint256"}],"anonymous":false},{"type":"error","name":"ContractFrozen","inputs":[]},{"type":"error","name":"DataCommitmentNotFound","inputs":[]},{"type":"error","name":"LatestHeaderNotFound","inputs":[]},{"type":"error","name":"TargetBlockNotInRange","inputs":[]},{"type":"error","name":"TrustedBlockMismatch","inputs":[]},{"type":"error","name":"TrustedHeaderNotFound","inputs":[]}]`

	// Inflation params
	InflationRateChange = sdkmath.LegacyMustNewDecFromStr("0.13")
	InflationMax        = sdkmath.LegacyMustNewDecFromStr("0.2")
	InflationMin        = sdkmath.LegacyMustNewDecFromStr("0.07")
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

	// Ethereum
	ethResource *dockertest.Resource

	// Sequencer
	valResources []*dockertest.Resource

	// SuccinctX
	succinctOperatorResource *dockertest.Resource
	succinctRelayerResource  *dockertest.Resource

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
	err = s.WaitForBlocks(s.Ctx(), 1, time.Minute)
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
	if s.succinctOperatorResource != nil {
		_ = s.dockerPool.Purge(s.succinctOperatorResource)
	}
	if s.succinctRelayerResource != nil {
		_ = s.dockerPool.Purge(s.succinctRelayerResource)
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
		2*time.Second,
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
