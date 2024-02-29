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
		genesisState, err = simtestutil.GenesisStateWithValSet(
			app.AppCodec(),
			genesisState,
			valSet,
			[]authtypes.GenesisAccount{acc},
			balances...,
		)
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
