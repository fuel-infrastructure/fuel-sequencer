package authorize_transactions_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	e2etestsuite "github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	minttypes "github.com/fuel-infrastructure/fuel-sequencer/x/mint/types"
	"github.com/stretchr/testify/suite"
)

type AuthorizeTransactionsTestSuite struct {
	e2etestsuite.E2ETestSuite
}

func TestAuthorizeTransactionsTestSuite(t *testing.T) {
	suite.Run(t, new(AuthorizeTransactionsTestSuite))
}

// SetupTest modifies the genesis file as required by AuthorizeTransactionsTestSuite
func (s *AuthorizeTransactionsTestSuite) SetupTest() {

	genesisModifier := e2etestsuite.ModifyGenesisFunc(
		func(cdc codec.Codec, genesisState map[string]json.RawMessage) error {

			// ------ Add a high SupplyDeltaPeriod so that we don't have to worry about supply delta messages

			var bridgeGenState bridgetypes.GenesisState
			s.Require().NoError(cdc.UnmarshalJSON(genesisState[bridgetypes.ModuleName], &bridgeGenState))

			bridgeGenState.Params.SupplyDeltaPeriod = uint64(1000)

			bz, err := cdc.MarshalJSON(&bridgeGenState)
			s.Require().NoError(err)
			genesisState[bridgetypes.ModuleName] = bz

			// ----- Define accounts on the Sequencer for the test Ethereum addresses so that we can execute authorized
			// Transactions

			var accountsGenState authtypes.GenesisState
			s.Require().NoError(cdc.UnmarshalJSON(genesisState[authtypes.ModuleName], &accountsGenState))

			accs, err := authtypes.UnpackAccounts(accountsGenState.Accounts)
			if err != nil {
				panic(fmt.Errorf("failed to get accounts from any: %w", err))
			}

			for _, address := range e2etestsuite.ETH_ADDRESS_SEQ {
				baseAccount := authtypes.NewBaseAccount(sdk.MustAccAddressFromBech32(address), nil, 0, 0)
				ethOwnedBaseAccount := bridgetypes.NewEthOwnedBaseAccount(baseAccount, address)
				accs = append(accs, ethOwnedBaseAccount)
			}

			accs = authtypes.SanitizeGenesisAccounts(accs)

			genAccs, err := authtypes.PackAccounts(accs)
			if err != nil {
				panic(fmt.Errorf("failed to convert accounts into any's: %w", err))
			}

			accountsGenState.Accounts = genAccs

			bz, err = cdc.MarshalJSON(&accountsGenState)
			s.Require().NoError(err)
			genesisState[authtypes.ModuleName] = bz

			// ----- Define balances on the Sequencer for the test Ethereum addresses so that we can execute authorized
			// Transactions

			var bankGenState banktypes.GenesisState
			s.Require().NoError(cdc.UnmarshalJSON(genesisState[banktypes.ModuleName], &bankGenState))

			for _, address := range e2etestsuite.ETH_ADDRESS_SEQ {
				balances := banktypes.Balance{Address: address, Coins: sdk.NewCoins(e2etestsuite.InitBalanceCoin)}
				bankGenState.Balances = append(bankGenState.Balances, balances)
				bankGenState.Supply = bankGenState.Supply.Add(balances.Coins...)
			}

			bankGenState.Balances = banktypes.SanitizeGenesisBalances(bankGenState.Balances)

			bz, err = cdc.MarshalJSON(&bankGenState)
			s.Require().NoError(err)
			genesisState[banktypes.ModuleName] = bz

			// ----- Set a non-zero inflation rate to generate staking rewards. This is required to test out an
			// authorized MsgWithdrawDelegatorReward

			var mintGenState minttypes.GenesisState
			s.Require().NoError(cdc.UnmarshalJSON(genesisState[minttypes.ModuleName], &mintGenState))

			mintGenState.Params.InflationRateChange = e2etestsuite.InflationRateChange
			mintGenState.Params.InflationMax = e2etestsuite.InflationMax
			mintGenState.Params.InflationMin = e2etestsuite.InflationMin

			bz, err = cdc.MarshalJSON(&mintGenState)
			s.Require().NoError(err)
			genesisState[minttypes.ModuleName] = bz

			// ----- Increase the voting period substantially to allow for authorized MsgVote to go through comfortably
			var govGenState govtypesv1.GenesisState
			s.Require().NoError(cdc.UnmarshalJSON(genesisState[govtypes.ModuleName], &govGenState))

			oneHour := time.Hour
			govGenState.Params.VotingPeriod = &oneHour

			bz, err = cdc.MarshalJSON(&govGenState)
			s.Require().NoError(err)
			genesisState[govtypes.ModuleName] = bz

			return nil
		},
	)
	s.GenesisOverrides = &genesisModifier

	s.E2ETestSuite.SetupTest()
}
