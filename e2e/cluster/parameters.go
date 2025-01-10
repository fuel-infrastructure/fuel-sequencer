package cluster

import (
	"path/filepath"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
)

const (
	BridgeDenom = testsuite.BridgeDenom // as configuration is being generated from testsuite, tied to it - effectively constant
)

var (
	// Binary Parameters
	makefileDir string = "/home/user/fuel-sequencer"                                 // Absolute path to the directory where makefile is located
	wantArch    string = "linux-amd64"                                               // Arch specified from build binary suffix
	buildPath   string = makefileDir + "/build"                                      // Path where binary will be built
	dataDir     string = makefileDir + "/e2e/cluster/data"                           // Directory with template data
	servicePath string = makefileDir + "/e2e/cluster/systemd/fuelsequencerd.service" // Path to the systemd service file
	systemdPath string = "/etc/systemd/system/fuelsequencerd.service"                // Path to the systemd service file on the remote machine
	binaryName  string = "fuelsequencerd"                                            // Name of the binary

	// Sequencer Parameters
	chainName = "seq-benchnet-1"

	// Use predefined mnemonics for deterministic addresses
	// Also dictates the number of validators
	mnemonics = []string{
		"test test test test test test test test test test test junk",
		"dinner crash nurse casino baby fold race cheese elite column sausage sleep close royal rain over mechanic minimum outdoor conduct cash wagon frog evidence",
		"gaze drama excess raven follow antenna swallow beef upper myself question pitch course ill adult century crisp ice rough match praise sing unveil vintage",
		// "bar describe panda mosquito quiz room daring round nurse disagree swallow frown hat repeat recall flight skin sketch volume dutch range grunt assist nerve",
		// "bonus clinic owner choose grief soda ride divorce album oval tone mixed mechanic coin defense wonder tumble vault sorry great hover neither security amazing",
	}

	BridgeDenomTotalSupply, initSupplyValid = sdkmath.NewIntFromString("10000000000000000000") // 10 bil x 1e9

	// Balance and staked amount per validator
	initBalance, initBalanceValid = sdkmath.NewIntFromString("20000000000000000000") // 20 bil x 1e9
	initStaked, initStakedValid   = sdkmath.NewIntFromString("2000000000")           // 2e9
	InitBalanceCoin               = sdk.NewCoin(BridgeDenom, initBalance)
	InitStakedCoin                = sdk.NewCoin(BridgeDenom, initStaked)

	// Genesis configs
	governanceVotingPeriod = time.Second * 20 // default - can be overridden
	supplyDeltaPeriod      = uint64(10)       // default - can be overridden
	blobMaxBytes           = uint64(2147483648)
	blockMaxGas            = uint64(4294967296)

	// Gas configs
	minGasPrices = "0.0"

	// Remote Parameters
	destinations = []destination{
		{
			peer_ip: "127.0.0.1",
			host:    "localhost",
			user:    "benchmarks",
			pass:    "password",
			dir:     "/home/benchmarks",
		},
	}
)

func chainHomeDir(d destination) string {
	return filepath.Join(d.dir, ".fuelsequencer")
}

func remoteBinaryPath(d destination) string {
	return filepath.Join(d.dir, binaryName)
}

func init() {
	if !initSupplyValid {
		logging.Panicw("parameter initSupplyValid is invalid")
	}
	if !initBalanceValid {
		logging.Panicw("parameter initBalanceValid is invalid")
	}
	if !initStakedValid {
		logging.Panicw("parameter initStakedValid is invalid")
	}
}
