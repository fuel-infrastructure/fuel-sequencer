package withdrawals_test

import (
	"encoding/json"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/stretchr/testify/suite"

	e2etestsuite "github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

type WithdrawalsTestSuite struct {
	e2etestsuite.E2ETestSuite
}

func TestWithdrawalsTestSuite(t *testing.T) {
	suite.Run(t, new(WithdrawalsTestSuite))
}

// SetupTest sets a high supply delta period so that we can really focus on Ethereum events.
func (s *WithdrawalsTestSuite) SetupTest() {

	setHighSupplyDeltaPeriod := e2etestsuite.ModifyGenesisFunc(
		func(cdc codec.Codec, genesisState map[string]json.RawMessage) error {

			var bridgeGenState bridgetypes.GenesisState
			s.Require().NoError(cdc.UnmarshalJSON(genesisState[bridgetypes.ModuleName], &bridgeGenState))

			bridgeGenState.Params.SupplyDeltaPeriod = uint64(1000)

			bz, err := cdc.MarshalJSON(&bridgeGenState)
			s.Require().NoError(err)
			genesisState[bridgetypes.ModuleName] = bz

			return nil
		},
	)
	s.GenesisOverrides = &setHighSupplyDeltaPeriod

	s.E2ETestSuite.SetupTest()
}
