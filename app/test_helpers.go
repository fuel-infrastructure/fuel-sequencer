package app

import (
	"encoding/json"

	"cosmossdk.io/log"
	"cosmossdk.io/math"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtypes "github.com/cometbft/cometbft/types"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	TestAppStartingHeight = 1
	TestAppChainID        = "fuelsequencer-1"
)

// SetupTestingApp initializes a new FuelSequencerApp
// Note: use NewTMLogger(NewSyncWriter(Stdout)) instead of NewNopLogger if you want to see test logs
func SetupTestingApp(isCheckTx bool) *FuelSequencerApp {
	InitSDKConfig()
	InitCometBFTConfig()
	InitAppConfig()

	db := dbm.NewMemDB()
	appOpts := simtestutil.AppOptionsMap{
		flags.FlagInitHeight: TestAppStartingHeight,
		flags.FlagChainID:    TestAppChainID,
	}
	app, err := NewFuelSequencerApp(
		log.NewNopLogger(),
		db,
		nil,
		true,
		appOpts,
	)
	if err != nil {
		panic(err)
	}
	if !isCheckTx {
		_, _ = app.BaseApp.InitChain(
			&abci.RequestInitChain{
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: simtestutil.DefaultConsensusParams,
				AppStateBytes:   getDefaultGenesisStateBytes(app),
			},
		)
	}

	return app
}

var defaultGenesisBz []byte

func getDefaultGenesisStateBytes(app *FuelSequencerApp) []byte {
	if len(defaultGenesisBz) == 0 {
		privVal := mock.NewPV()
		pubKey, err := privVal.GetPubKey()
		if err != nil {
			panic(err)
		}
		// create validator set with single validator
		validator := cmtypes.NewValidator(pubKey, 1)
		valSet := cmtypes.NewValidatorSet([]*cmtypes.Validator{validator})

		// generate genesis account
		senderPrivKey := secp256k1.GenPrivKey()
		acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
		balances := []banktypes.Balance{
			{
				Address: acc.GetAddress().String(),
				Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100000000000000))),
			},
		}

		genesisState := app.DefaultGenesis()
		genesisState, err = simtestutil.GenesisStateWithValSet(app.AppCodec(), genesisState, valSet, []authtypes.GenesisAccount{acc}, balances...)
		if err != nil {
			panic(err)
		}
		stateBytes, err := json.MarshalIndent(genesisState, "", " ")
		if err != nil {
			panic(err)
		}
		defaultGenesisBz = stateBytes
	}
	return defaultGenesisBz
}

//func genesisStateWithValSet(
//	app *FuelSequencerApp, genesisState GenesisState,
//	valSet *cmtypes.ValidatorSet, genAccs []authtypes.GenesisAccount,
//	balances ...banktypes.Balance,
//) GenesisState {
//	// set genesis accounts
//	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
//	genesisState[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(authGenesis)
//
//	validators := make([]stakingtypes.Validator, 0, len(valSet.Validators))
//	delegations := make([]stakingtypes.Delegation, 0, len(valSet.Validators))
//
//	bondAmt := sdk.DefaultPowerReduction
//	oneDec := math.LegacyOneDec()
//
//	for _, val := range valSet.Validators {
//		pk, err := cryptocodec.FromTmPubKeyInterface(val.PubKey)
//		if err != nil {
//			panic(err)
//		}
//		pkAny, err := cdctypes.NewAnyWithValue(pk)
//		if err != nil {
//			panic(err)
//		}
//		validator := stakingtypes.Validator{
//			OperatorAddress:   sdk.ValAddress(val.Address).String(),
//			ConsensusPubkey:   pkAny,
//			Jailed:            false,
//			Status:            stakingtypes.Bonded,
//			Tokens:            bondAmt,
//			DelegatorShares:   math.LegacyOneDec(),
//			Description:       stakingtypes.Description{},
//			UnbondingHeight:   int64(0),
//			UnbondingTime:     time.Unix(0, 0).UTC(),
//			Commission:        stakingtypes.NewCommission(oneDec, oneDec, oneDec),
//			MinSelfDelegation: math.ZeroInt(),
//		}
//		validators = append(validators, validator)
//		delegation := stakingtypes.NewDelegation(genAccs[0].GetAddress().String(), val.Address.String(), oneDec)
//		delegations = append(delegations, delegation)
//
//	}
//	// set validators and delegations
//	stakingGenesis := stakingtypes.NewGenesisState(stakingtypes.DefaultParams(), validators, delegations)
//	genesisState[stakingtypes.ModuleName] = app.AppCodec().MustMarshalJSON(stakingGenesis)
//
//	totalSupply := sdk.NewCoins()
//	for _, b := range balances {
//		// add genesis acc tokens to total supply
//		totalSupply = totalSupply.Add(b.Coins...)
//	}
//
//	for range delegations {
//		// add delegated tokens to total supply
//		totalSupply = totalSupply.Add(sdk.NewCoin(sdk.DefaultBondDenom, bondAmt))
//	}
//
//	// add bonded amount to bonded pool module account
//	balances = append(balances, banktypes.Balance{
//		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
//		Coins:   sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, bondAmt)},
//	})
//
//	// update total supply
//	bankGenesis := banktypes.NewGenesisState(banktypes.DefaultGenesisState().Params, balances, totalSupply,
//		[]banktypes.Metadata{}, []banktypes.SendEnabled{})
//	genesisState[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)
//
//	return genesisState
//}
