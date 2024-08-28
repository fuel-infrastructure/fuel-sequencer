package full_test

import (
	"encoding/json"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/stretchr/testify/suite"

	e2etestsuite "github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

type FullTestSuite struct {
	e2etestsuite.E2ETestSuite
}

func TestFullTestSuite(t *testing.T) {
	suite.Run(t, new(FullTestSuite))
}

// SetupTest sets a high supply delta period so that we can really focus on Ethereum events.
func (s *FullTestSuite) SetupTest() {

	setLowSupplyDeltaPeriod := e2etestsuite.ModifyGenesisFunc(
		func(cdc codec.Codec, genesisState map[string]json.RawMessage) error {

			var bridgeGenState bridgetypes.GenesisState
			s.Require().NoError(cdc.UnmarshalJSON(genesisState[bridgetypes.ModuleName], &bridgeGenState))

			bridgeGenState.Params.SupplyDeltaPeriod = uint64(3)

			bz, err := cdc.MarshalJSON(&bridgeGenState)
			s.Require().NoError(err)
			genesisState[bridgetypes.ModuleName] = bz

			return nil
		},
	)
	s.GenesisOverrides = &setLowSupplyDeltaPeriod

	s.E2ETestSuite.SetupTest()
}
