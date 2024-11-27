package types_test

import (
	"testing"

	sdk "cosmossdk.io/math"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"github.com/stretchr/testify/require"
)

func TestSupplyDeltaInfo_ValidateBasic(t *testing.T) {
	validLastSupply := sdk.NewInt(100_000_000_000)
	validOffset := sdk.NewInt(650_000)
	validToReport := sdk.NewInt(-320_000_000)

	tests := []struct {
		desc            string
		supplyDeltaInfo types.SupplyDeltaInfo
		valid           bool
	}{
		{
			desc:            "default is valid",
			supplyDeltaInfo: *types.DefaultGenesis().SupplyDeltaInfo,
			valid:           true,
		},
		{
			desc: "non-default is valid",
			supplyDeltaInfo: types.SupplyDeltaInfo{
				LastSupply: validLastSupply,
				Offset:     validOffset,
				ToReport:   validToReport,
			},
			valid: true,
		},
		{
			desc: "invalid last supply",
			supplyDeltaInfo: types.SupplyDeltaInfo{
				LastSupply: sdk.NewInt(-1),
				Offset:     validOffset,
				ToReport:   validToReport,
			},
			valid: false,
		},
		// this line is used by starport scaffolding # types/genesis/testcase
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.supplyDeltaInfo.ValidateBasic()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
