// Package sequencer provides functionality for setting up and managing a distributed
// network of Fuel Sequencer validator nodes. It handles binary building, configuration,
// deployment and management of the network.
package sequencer

// TODO: This is a temporary file to store parameters for the cluster.
// It should be replaced with a proper configuration management system in the future.
// In the meantime, primary parameters to configure are:
// - makefileDir: The directory where the makefile is located
// - destinations: The list of destinations to deploy to
//
// The rest are applied remotely or intended to be consistent across a deployment. Ultimately depends on the usecase.

import (
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"go.uber.org/zap"

	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
)

const (
	bridgeDenom = testsuite.BridgeDenom // as configuration is being generated from testsuite, tied to it - effectively constant

	ChainName = "seq-benchnet-1"

	// Genesis configs
	governanceVotingPeriod = time.Second * 20 // default - can be overridden
	supplyDeltaPeriod      = uint64(10)       // default - can be overridden
	blobMaxBytes           = uint64(2147483648)
	blockMaxGas            = uint64(4294967296)
	mempoolMaxTxBytes      = int(4294967296)   // 4 GiB
	mempoolMaxTxsBytes     = int64(4294967296) // 4 GiB

	// Gas configs
	minGasPrices = "0.0"
)

var (
	// Use predefined Mnemonics for deterministic addresses
	// Also dictates the number of validators
	Mnemonics = []string{
		"test test test test test test test test test test test junk",
		"dinner crash nurse casino baby fold race cheese elite column sausage sleep close royal rain over mechanic minimum outdoor conduct cash wagon frog evidence",
		"gaze drama excess raven follow antenna swallow beef upper myself question pitch course ill adult century crisp ice rough match praise sing unveil vintage",
		// "bar describe panda mosquito quiz room daring round nurse disagree swallow frown hat repeat recall flight skin sketch volume dutch range grunt assist nerve",
		// "bonus clinic owner choose grief soda ride divorce album oval tone mixed mechanic coin defense wonder tumble vault sorry great hover neither security amazing",
	}

	bridgeDenomTotalSupply, initSupplyValid = sdkmath.NewIntFromString("10000000000000000000") // 10 bil x 1e9

	// Balance and staked amount per validator
	initBalance, initBalanceValid = sdkmath.NewIntFromString("20000000000000000000") // 20 bil x 1e9
	initStaked, initStakedValid   = sdkmath.NewIntFromString("2000000000")           // 2e9
	initBalanceCoin               = sdk.NewCoin(bridgeDenom, initBalance)
	initStakedCoin                = sdk.NewCoin(bridgeDenom, initStaked)
)

// CheckParameters validates the initialization parameters
func CheckParameters(logging *zap.SugaredLogger) {
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
