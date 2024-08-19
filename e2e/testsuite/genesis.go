package testsuite

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"cosmossdk.io/math"
	cmjson "github.com/cometbft/cometbft/libs/json"
	cmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

type ModifyGenesisFunc func(codec.Codec, map[string]json.RawMessage) error

func ChainedModifyGenesis(modifyGenesisFuncs ...ModifyGenesisFunc) ModifyGenesisFunc {
	return func(
		cdc codec.Codec, genesisState map[string]json.RawMessage,
	) (err error) {
		for i, mgf := range modifyGenesisFuncs {
			err = mgf(cdc, genesisState)
			if err != nil {
				panic(fmt.Sprintf("chained genesis modifier %d failed with error: %s", i, err.Error()))
			}
		}
		return nil
	}
}

func getGenDoc(path string) (*cmtypes.GenesisDoc, error) {
	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config
	config.SetRoot(path)

	genFile := config.GenesisFile()
	doc := &cmtypes.GenesisDoc{}

	if _, err := os.Stat(genFile); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else {
		var err error

		doc, err = cmtypes.GenesisDocFromFile(genFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read genesis doc from file: %w", err)
		}
	}

	return doc, nil
}

func addGenesisAccount(path, moniker, amountStr string, accAddr sdk.AccAddress) error { //nolint:unparam
	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(path)
	config.Moniker = moniker

	coins, err := sdk.ParseCoinsNormalized(amountStr)
	if err != nil {
		return fmt.Errorf("failed to parse coins: %w", err)
	}

	balances := banktypes.Balance{Address: accAddr.String(), Coins: coins.Sort()}
	genAccount := authtypes.NewBaseAccount(accAddr, nil, 0, 0)

	genFile := config.GenesisFile()
	appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
	if err != nil {
		return fmt.Errorf("failed to unmarshal genesis state: %w", err)
	}

	authGenState := authtypes.GetGenesisStateFromAppState(cdc, appState)

	accs, err := authtypes.UnpackAccounts(authGenState.Accounts)
	if err != nil {
		return fmt.Errorf("failed to get accounts from any: %w", err)
	}

	if accs.Contains(accAddr) {
		return fmt.Errorf("failed to add account to genesis state; account already exists: %s", accAddr)
	}

	// Add the new account to the set of genesis accounts and sanitize the
	// accounts afterwards.
	accs = append(accs, genAccount)
	accs = authtypes.SanitizeGenesisAccounts(accs)

	genAccs, err := authtypes.PackAccounts(accs)
	if err != nil {
		return fmt.Errorf("failed to convert accounts into any's: %w", err)
	}

	authGenState.Accounts = genAccs

	authGenStateBz, err := cdc.MarshalJSON(&authGenState)
	if err != nil {
		return fmt.Errorf("failed to marshal auth genesis state: %w", err)
	}

	appState[authtypes.ModuleName] = authGenStateBz

	bankGenState := banktypes.GetGenesisStateFromAppState(cdc, appState)
	bankGenState.Balances = append(bankGenState.Balances, balances)
	bankGenState.Balances = banktypes.SanitizeGenesisBalances(bankGenState.Balances)

	bankGenStateBz, err := cdc.MarshalJSON(bankGenState)
	if err != nil {
		return fmt.Errorf("failed to marshal bank genesis state: %w", err)
	}

	appState[banktypes.ModuleName] = bankGenStateBz

	appStateJSON, err := json.Marshal(appState)
	if err != nil {
		return fmt.Errorf("failed to marshal application genesis state: %w", err)
	}

	genDoc.AppState = appStateJSON
	return genutil.ExportGenesisFile(genDoc, genFile)
}

func (s *E2ETestSuite) initFuelSequencerGenesis() {
	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(s.Chain.validators[0].configDir())
	config.Moniker = s.Chain.validators[0].moniker

	genFilePath := config.GenesisFile()
	appGenState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFilePath)
	s.Require().NoError(err)

	// set short voting period to allow gov proposals in tests
	var govGenState govtypesv1.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[govtypes.ModuleName], &govGenState))
	votingPeriod := governanceVotingPeriod
	govGenState.Params.VotingPeriod = &votingPeriod
	govGenState.Params.MinDeposit = sdk.Coins{{Denom: BridgeDenom, Amount: math.OneInt()}}
	govGenState.Params.ExpeditedMinDeposit = sdk.Coins{{Denom: BridgeDenom, Amount: math.OneInt()}}
	bz, err := cdc.MarshalJSON(&govGenState)
	s.Require().NoError(err)
	appGenState[govtypes.ModuleName] = bz

	// set mint denom
	var mintGenState minttypes.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[minttypes.ModuleName], &mintGenState))
	mintGenState.Params.InflationMax = math.LegacyZeroDec()
	mintGenState.Params.InflationMin = math.LegacyZeroDec()
	mintGenState.Params.InflationRateChange = math.LegacyZeroDec()
	mintGenState.Minter.Inflation = math.LegacyZeroDec()
	bz, err = cdc.MarshalJSON(&mintGenState)
	s.Require().NoError(err)
	appGenState[minttypes.ModuleName] = bz

	// TODO: genesis supply will be incorrect if we add more accounts
	var bankGenState banktypes.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[banktypes.ModuleName], &bankGenState))
	genesisSupply := int64(len(s.Chain.validators) * initBalance)
	bankGenState.Supply = sdk.NewCoins(sdk.NewCoin(BridgeDenom, math.NewInt(genesisSupply)))
	bz, err = cdc.MarshalJSON(&bankGenState)
	s.Require().NoError(err)
	appGenState[banktypes.ModuleName] = bz

	vestingStartingTime, err := time.Parse(time.DateOnly, "2024-01-01")
	s.Require().NoError(err)
	ethBlockNumber, err := s.Chain.ethClient.BlockNumber(s.Ctx()) // start syncing from the current Ethereum block
	s.Require().NoError(err)
	s.T().Logf("set last Ethereum block synced to %d", ethBlockNumber)

	var bridgeGenState bridgetypes.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[bridgetypes.ModuleName], &bridgeGenState))
	bridgeGenState.Params.BridgeDenom = BridgeDenom
	bridgeGenState.Params.SupplyDeltaPeriod = supplyDeltaPeriod
	bridgeGenState.Params.VestingStartTime = vestingStartingTime
	bridgeGenState.Params.BridgeDenomTotalSupply = math.NewInt(BridgeDenomTotalSupply)
	bridgeGenState.LastEthereumBlockSynced = ethBlockNumber
	bz, err = cdc.MarshalJSON(&bridgeGenState)
	s.Require().NoError(err)
	appGenState[bridgetypes.ModuleName] = bz

	var genUtilGenState genutiltypes.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[genutiltypes.ModuleName], &genUtilGenState))

	// generate genesis txs
	genTxs := make([]json.RawMessage, len(s.Chain.validators))
	for i, val := range s.Chain.validators {
		createValmsg, err := val.buildCreateValidatorMsg(InitStakedCoin)
		s.Require().NoError(err)

		signedTx, err := val.signMsg(createValmsg)
		s.Require().NoError(err)

		txRaw, err := cdc.MarshalJSON(signedTx)
		s.Require().NoError(err)

		genTxs[i] = txRaw
	}

	genUtilGenState.GenTxs = genTxs

	bz, err = cdc.MarshalJSON(&genUtilGenState)
	s.Require().NoError(err)
	appGenState[genutiltypes.ModuleName] = bz

	// Apply any genesis overrides
	if s.GenesisOverrides != nil {
		err = (*s.GenesisOverrides)(cdc, appGenState)
		s.Require().NoError(err)
	}

	// serialize genesis state
	bz, err = json.MarshalIndent(appGenState, "", "  ")
	s.Require().NoError(err)

	genDoc.AppState = bz

	bz, err = cmjson.MarshalIndent(genDoc, "", "  ")
	s.Require().NoError(err)

	// write the updated genesis file to each validator
	for _, val := range s.Chain.validators {
		s.Require().NoError(writeFile(filepath.Join(val.configDir(), "config", "genesis.json"), bz))
	}
}
