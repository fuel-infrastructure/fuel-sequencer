package basic_test

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/codec"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	e2etestsuite "github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	"github.com/stretchr/testify/suite"
	"testing"
)

type BasicTestSuite struct {
	e2etestsuite.E2ETestSuite
}

func TestBasicTestSuite(t *testing.T) {
	suite.Run(t, new(BasicTestSuite))
}

func (s *BasicTestSuite) SetupTest() {

	genesisModifier := e2etestsuite.ModifyGenesisFunc(
		func(cdc codec.Codec, genesisState map[string]json.RawMessage) error {

			// ----- Set a non-zero inflation rate to generate staking rewards

			var mintGenState minttypes.GenesisState
			s.Require().NoError(cdc.UnmarshalJSON(genesisState[minttypes.ModuleName], &mintGenState))

			mintGenState.Params.InflationRateChange = e2etestsuite.InflationRateChange
			mintGenState.Params.InflationMax = e2etestsuite.InflationMax
			mintGenState.Params.InflationMin = e2etestsuite.InflationMin
			mintGenState.Params.GoalBonded = e2etestsuite.GoalBonded
			mintGenState.Params.BlocksPerYear = e2etestsuite.BlocksPerYear

			bz, err := cdc.MarshalJSON(&mintGenState)
			s.Require().NoError(err)
			genesisState[minttypes.ModuleName] = bz

			return nil
		},
	)
	s.GenesisOverrides = &genesisModifier

	s.E2ETestSuite.SetupTest()
}
