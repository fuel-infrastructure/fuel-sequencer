package types_test

import (
	"testing"

	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"
	"github.com/stretchr/testify/require"
)

func TestSlashReport_ValidateBasic(t *testing.T) {
	tests := []struct {
		name        string
		slashReport types.SlashReport
		expErrMsg   string
	}{
		{
			name:        "valid SlashReport - more than one SlashEntry",
			slashReport: testtypes.ValidSlashReport1,
		},
		{
			name:        "valid SlashReport - one SlashEntry",
			slashReport: testtypes.ValidSlashReport2,
		},
		{
			name:        "invalid SlashReport - height cannot be zero",
			slashReport: testtypes.InvalidSlashReportHeightZero,
			expErrMsg:   "slash report height cannot be zero",
		},
		{
			name:        "invalid SlashReport - entries cannot be empty",
			slashReport: testtypes.InvalidSlashReportEmptyEntries,
			expErrMsg:   "entries cannot be empty",
		},
		{
			name:        "invalid SlashReport - entries cannot be nil",
			slashReport: testtypes.InvalidSlashReportNilEntries,
			expErrMsg:   "entries cannot be empty",
		},
		{
			name:        "invalid SlashReport - entries cannot be invalid",
			slashReport: testtypes.InvalidSlashReportValidatorAddressNotValoper,
			expErrMsg:   "ValidatorAddress is not a valid Bech32 operator address",
		},
		{
			name:        "invalid SlashReport - slash entries with non unique keys",
			slashReport: testtypes.InvalidSlashReportNonUniqueSlashEntries,
			expErrMsg:   "duplicate slash entry found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.slashReport.ValidateBasic()
			if len(tc.expErrMsg) > 0 {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
