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
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"go.uber.org/zap"

	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/pkg/setup"
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

	// Additional non-validator genesis accounts with deterministic mnemonics.
	// These accounts will have balances at genesis but will not be validators.
	// You can add or remove mnemonics from this list as needed - the system will
	// automatically create one account per mnemonic with the configured balance.
	// To add more accounts, use: go run e2e/cluster/scripts/generate_mnemonics.go -count N
	// Currently configured: 60 accounts (all valid BIP-39 mnemonics)
	AdditionalGenesisMnemonics = []string{
		"shock rail nothing gloom run kangaroo forget affair they remove company fatal",
		"misery muscle brown render drink evolve latin spy mention guitar spell fever",
		"ancient model beef staff hole volcano job gas scan skirt kid tooth",
		"flock crop cupboard fortune venue excuse antique journey tongue script addict flag",
		"metal lounge ring grocery obtain mansion derive icon employ tortoise review exotic",
		"wedding acid sail broken run exile unfair april anchor make settle weekend",
		"achieve kid reject high damage aspect afraid cactus plate medal hazard cram",
		"race before oxygen will detail shiver brain faith beef marriage odor loud",
		"armed error capable film income crucial toilet curious uncle sun wink moral",
		"afford school swallow inflict visual reform verify mouse immune medal pill culture",
		"pyramid fiscal half raw olive learn payment buyer they first bamboo false",
		"catalog priority jump demand bubble panic chunk gorilla rail chalk soap simple",
		"river chronic fox wash team afford marine oyster feature inject love tip",
		"acquire sample enter crew boss hockey afford salon find roof toe web",
		"bleak peanut glimpse relief modify earn paddle sudden crucial neutral tube rifle",
		"brand injury kind mushroom bullet music water slogan spatial humor start keen",
		"board process dinner pen disagree country fall voice tail vague yellow produce",
		"pilot primary ring pipe frog delay person team alter dinner club awful",
		"sample giant play aim flock lady weapon lesson country pitch episode ranch",
		"analyst marriage ankle balance finish extra bomb aerobic suspect pipe zoo solve",
		"race wear lunch physical drive fever gun forget first entire hunt wait",
		"food real inform nuclear roof true public surface burger amount slow roast",
		"usual skin antique original claw crazy amazing guide body rack cram defense",
		"coin south aim alcohol chaos bag autumn more increase traffic green hungry",
		"afraid traffic time attend caught little cabbage job attitude tone cloud spare",
		"express print rocket explain undo end throw year present few venue cabbage",
		"alpha power crunch quiz jump weird video winter indicate buddy awkward option",
		"train lawsuit orphan foster risk biology leader ancient minimum victory suffer panel",
		"replace truly claw marriage resource crazy skirt garden magic retreat marine parade",
		"crew jazz oval off abstract border rely civil delay repeat office kiwi",
		"chair curious stay parrot piece swallow intact arena horror crawl sibling forget",
		"purity family document possible grab pipe range rifle always cat mechanic copy",
		"capable recipe hip travel damp proof that skate note pupil burst mercy",
		"armor blade chief symptom couch author dismiss family library endorse aware half",
		"nature crawl stand credit absurd announce fold mercy spider run thrive frog",
		"rich fiction trigger regret okay success unlock thank sea celery slight warm",
		"monitor jazz expand chase tip actual memory depend fuel upgrade soldier salad",
		"render across favorite power muffin sea wealth shell castle split fence apple",
		"stable wedding dinner ostrich barely object juice fringe near spare fork fiber",
		"eager science bachelor mind welcome design bright feel crowd error black hammer",
		"divorce tongue retreat ritual surface current spice chalk print loud arm either",
		"rhythm wrestle country shoot ahead rabbit tongue body comic pulse sorry pumpkin",
		"combine actor badge like width ghost dash slogan spoon orphan invite rocket",
		"change yellow cook drift nurse ahead click orchard rude tube velvet satisfy",
		"feed submit album symptom wall exhibit view avocado maximum claw frost seat",
		"funny catch misery omit bar thank lens until strong renew bridge odor",
		"must assist raw rely desert vapor success burger travel frequent grief member",
		"tiny crew put glue fee inflict despair twin few man fiscal when",
		"creek unfair mosquito spy era subway old glue clown popular dinner cart",
		"build entry work utility cube nurse already final sand firm vendor pulp",
		"regret vast virtual baby rather destroy measure ritual fit simple travel high",
		"puzzle rabbit scale seven spend radar festival pretty symbol miracle giant exhaust",
		"town maximum puppy follow path climb idea check shrug kitten another apology",
		"tomorrow black hungry hope tell violin attract clever plug share clay horse",
		"hobby cube dynamic afraid pride relax spatial supply sting february fortune boy",
		"donor surge clip robust seek scare outer bag woman wife dismiss super",
		"soup vicious fold thumb media patch hunt path south alley wash prefer",
		"road mother danger tenant blind afraid balance churn base asset ginger hope",
		"dizzy fork alpha time idea salute wonder nasty fog hen metal churn",
		"sunny tumble dinosaur ramp envelope traffic unfair rich glue marble fragile display",
	}

	bridgeDenomTotalSupply, initSupplyValid = sdkmath.NewIntFromString("10000000000000000000") // 10 bil x 1e9

	// Balance and staked amount per validator
	initBalance, initBalanceValid = sdkmath.NewIntFromString("20000000000000000000") // 20 bil x 1e9
	initStaked, initStakedValid   = sdkmath.NewIntFromString("2000000000")           // 2e9
	initBalanceCoin               = sdk.NewCoin(bridgeDenom, initBalance)
	initStakedCoin                = sdk.NewCoin(bridgeDenom, initStaked)

	// Balance for additional genesis accounts (non-validators)
	additionalAccountBalance, additionalAccountBalanceValid = sdkmath.NewIntFromString("1000000000000000000") // 1 bil x 1e9
	additionalAccountBalanceCoin                            = sdk.NewCoin(bridgeDenom, additionalAccountBalance)
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
	if !additionalAccountBalanceValid {
		logging.Panicw("parameter additionalAccountBalanceValid is invalid")
	}

	// Log the number of additional genesis accounts that will be created
	logging.Infow("additional genesis accounts configured", "count", len(AdditionalGenesisMnemonics))
}

func CheckNetworkMnemonics(logging *zap.SugaredLogger) ([]setup.System, error) {
	systems := setup.Systems()
	totalInstances := setup.TotalInstances(systems)
	lm := len(Mnemonics)
	if totalInstances > lm {
		err := fmt.Errorf("check config: total number of instances (%d) exceeds number of mnemonics (%d)", totalInstances, lm)
		logging.Fatalw(err.Error())
		return nil, err
	}

	return systems, nil
}
