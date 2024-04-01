package apptesting

import (
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	"cosmossdk.io/store/rootmulti"

	tmtypes "github.com/cometbft/cometbft/proto/tendermint/types"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/fuel-infrastructure/fuel-sequencer/app"

	"github.com/stretchr/testify/suite"
)

type SuitelessAppTestHelper struct {
	App *app.FuelSequencerApp
	Ctx sdk.Context
}

type KeeperTestHelper struct {
	suite.Suite

	App *app.FuelSequencerApp

	QueryHelper *baseapp.QueryServiceTestHelper
	TestAccs    []sdk.AccAddress
}

// Setup sets up basic environment for suite (App, Ctx, and test accounts) with Now() as block time
func (s *KeeperTestHelper) Setup() {
	s.App = SetupTestingApp(false)
	s.QueryHelper = &baseapp.QueryServiceTestHelper{
		GRPCQueryRouter: s.App.GRPCQueryRouter(),
		Ctx:             s.Ctx().WithBlockTime(time.Now().UTC()),
	}
	s.TestAccs = sims.CreateRandomAccounts(3)
}

// Ctx dynamically gets the context of the chain
func (s *KeeperTestHelper) Ctx() sdk.Context {

	// Return a mock context
	return s.App.BaseApp.NewContext(false)
}

// CreateTestContext creates a test context.
func (s *KeeperTestHelper) CreateTestContext() sdk.Context {
	ctx, _ := s.CreateTestContextWithMultiStore()
	return ctx
}

// CreateTestContextWithMultiStore creates a test context and returns it together with multi store.
func (s *KeeperTestHelper) CreateTestContextWithMultiStore() (sdk.Context, store.CommitMultiStore) {
	db := dbm.NewMemDB()
	logger := log.NewNopLogger()

	ms := rootmulti.NewStore(db, logger, metrics.NoOpMetrics{})

	return sdk.NewContext(ms, tmtypes.Header{}, false, logger), ms
}

// FundAcc funds target address with specified amount.
func (s *KeeperTestHelper) FundAcc(ctx sdk.Context, acc sdk.AccAddress, amounts sdk.Coins) {
	err := testutil.FundAccount(ctx, s.App.BankKeeper, acc, amounts)
	s.Require().NoError(err)
}

// FundModuleAcc funds target modules with specified amount.
func (s *KeeperTestHelper) FundModuleAcc(ctx sdk.Context, moduleName string, amounts sdk.Coins) {
	err := testutil.FundModuleAccount(ctx, s.App.BankKeeper, moduleName, amounts)
	s.Require().NoError(err)
}

func (s *KeeperTestHelper) MintCoins(coins sdk.Coins) {
	err := s.App.BankKeeper.MintCoins(s.Ctx(), minttypes.ModuleName, coins)
	s.Require().NoError(err)
}

// EndBlock ends the block.
func (s *KeeperTestHelper) EndBlock() (sdk.EndBlock, error) {
	return s.App.EndBlocker(s.Ctx())
}

// AllocateRewardsToValidator allocates reward tokens to a distribution module then allocates rewards to the validator address.
func (s *KeeperTestHelper) AllocateRewardsToValidator(valAddr sdk.ValAddress, rewardAmt math.Int) {
	validator, err := s.App.StakingKeeper.GetValidator(s.Ctx(), valAddr)
	s.Require().NoError(err)

	// allocate reward tokens to distribution module
	coins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, rewardAmt)}
	err = testutil.FundModuleAccount(s.Ctx(), s.App.BankKeeper, distrtypes.ModuleName, coins)
	s.Require().NoError(err)

	// allocate rewards to validator
	decTokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(20000)}}
	err = s.App.DistrKeeper.AllocateTokensToValidator(s.Ctx(), validator, decTokens)
	s.Require().NoError(err)
}

// BuildTx builds a transaction.
func (s *KeeperTestHelper) BuildTx(
	txBuilder client.TxBuilder,
	msgs []sdk.Msg,
	sigV2 signing.SignatureV2,
	memo string, txFee sdk.Coins,
	gasLimit uint64,
) authsigning.Tx {
	err := txBuilder.SetMsgs(msgs[0])
	s.Require().NoError(err)

	err = txBuilder.SetSignatures(sigV2)
	s.Require().NoError(err)

	txBuilder.SetMemo(memo)
	txBuilder.SetFeeAmount(txFee)
	txBuilder.SetGasLimit(gasLimit)

	return txBuilder.GetTx()
}

// FastForwardBlocks is a helper function that increments the chain height
func (s *KeeperTestHelper) FastForwardBlocks(blocks int) error {
	for i := 0; i < blocks; i++ {
		_, err := s.EndBlock()
		if err != nil {
			return err
		}
	}
	return nil
}
