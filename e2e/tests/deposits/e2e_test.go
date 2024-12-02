package deposits_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	e2etestsuite "github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	"github.com/stretchr/testify/suite"
)

type DepositsTestSuite struct {
	e2etestsuite.E2ETestSuite
}

func TestDepositsTestSuite(t *testing.T) {
	suite.Run(t, new(DepositsTestSuite))
}

func (s *DepositsTestSuite) SetupTest() {

	genesisModifier := e2etestsuite.ModifyGenesisFunc(
		func(cdc codec.Codec, genesisState map[string]json.RawMessage) error {

			// ----- Set unbonding time to a second so that unbonding actions can be detected after 1 block

			var stakingGenState stakingtypes.GenesisState
			s.Require().NoError(cdc.UnmarshalJSON(genesisState[stakingtypes.ModuleName], &stakingGenState))

			stakingGenState.Params.UnbondingTime = time.Second // unbonding happens almost immediately (0 not accepted)

			bz, err := cdc.MarshalJSON(&stakingGenState)
			s.Require().NoError(err)
			genesisState[stakingtypes.ModuleName] = bz

			return nil
		},
	)
	s.GenesisOverrides = &genesisModifier

	s.E2ETestSuite.SetupTest()
}
