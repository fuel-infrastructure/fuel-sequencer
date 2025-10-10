// Package cluster provides functionality for setting up and managing a distributed
// network of Fuel Sequencer validator nodes. It handles binary building, configuration,
// deployment and management of the network.
package cluster

// TODO: This is a temporary file to store parameters for the cluster.
// It should be replaced with a proper configuration management system in the future.
// In the meantime, primary parameters to configure are:
// - makefileDir: The directory where the makefile is located
// - destinations: The list of destinations to deploy to
//
// The rest are applied remotely or intended to be consistent across a deployment. Ultimately depends on the usecase.

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
	makefileDir = "/home/user/fuel-sequencer"                                 // Absolute path to the directory where makefile is located
	wantArch    = "linux-amd64"                                               // Arch specified from build binary suffix
	buildPath   = makefileDir + "/build"                                      // Path where binary will be built
	dataDir     = makefileDir + "/e2e/cluster/data"                           // Directory with template data
	servicePath = makefileDir + "/e2e/cluster/systemd/fuelsequencerd.service" // Path to the systemd service file
	systemdPath = "/etc/systemd/system/fuelsequencerd.service"                // Path to the systemd service file on the remote machine
	remoteBinaryName = "fuelsequencerd"                                            // Name of the binary on remote

	// Blob Parameters
	blobComposePath = makefileDir + "/e2e/cluster/blob/docker-compose.blobpool.yml" // Path to the docker compose file
	blobRedisPath   = makefileDir + "/e2e/cluster/blob/redis.conf"                  // Path to the Redis configuration file

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
	mempoolMaxTxBytes      = int(4294967296)   // 4 GiB
	mempoolMaxTxsBytes     = int64(4294967296) // 4 GiB

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

// chainHomeDir returns the path to the chain's home directory on the remote host
func chainHomeDir(d destination) string {
	return filepath.Join(d.dir, ".fuelsequencer")
}

// remoteBinaryPath returns the path where the binary should be installed on the remote host
func remoteBinaryPath(d destination) string {
	return filepath.Join(d.dir, remoteBinaryName)
}

// remoteBlobDir returns the path where blob files should be stored on the remote host
func remoteBlobDir(d destination) string {
	return filepath.Join(d.dir, "blob")
}

// remoteBlobpoolComposePath returns the path where the docker compose file should be stored on the remote host
func remoteBlobpoolComposePath(d destination) string {
	return filepath.Join(remoteBlobDir(d), "docker-compose.blobpool.yml")
}

// remoteBlobStorageRedisConfPath returns the path where the Redis config file should be stored on the remote host
func remoteBlobStorageRedisConfPath(d destination) string {
	return filepath.Join(remoteBlobDir(d), "redis.conf")
}

// init validates the initialization parameters
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
