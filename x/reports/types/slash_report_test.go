package types_test

import (
	"testing"

	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"
	"github.com/stretchr/testify/require"
)

func TestValidateBasic(t *testing.T) {
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
			name:        "valid SlashReport - SlashEntry with zero DelegatorBondedBalance and DelegatorUnbondingBalance",
			slashReport: testtypes.ValidSlashReport3,
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
			name:        "invalid SlashReport - entries cannot have a non-valoper validator address",
			slashReport: testtypes.InvalidSlashReportValidatorAddressNotValoper,
			expErrMsg:   "ValidatorAddress is not a valid Bech32 Operator Address",
		},
		{
			name:        "invalid SlashReport - entries cannot have a non-account delegator address",
			slashReport: testtypes.InvalidSlashReportDelegatorAddressNotAccAddress,
			expErrMsg:   "DelegatorAddress is not a valid Bech32 Account address",
		},
		{
			name:        "invalid SlashReport - DelegatorSlashAmount cannot be negative",
			slashReport: testtypes.InvalidSlashReportDelegatorSlashAmountNegative,
			expErrMsg:   "DelegatorSlashAmount is not positive",
		},
		{
			name:        "invalid SlashReport - DelegatorSlashAmount cannot be zero",
			slashReport: testtypes.InvalidSlashReportDelegatorSlashAmountZero,
			expErrMsg:   "DelegatorSlashAmount is not positive",
		},
		{
			name:        "invalid SlashReport - DelegatorBondedBalance cannot be negative",
			slashReport: testtypes.InvalidSlashReportDelegatorBondedBalanceNegative,
			expErrMsg:   "DelegatorBondedBalance cannot be negative",
		},
		{
			name:        "invalid SlashReport - DelegatorUnbondingBalance cannot be negative",
			slashReport: testtypes.InvalidSlashReportDelegatorUnbondingBalanceNegative,
			expErrMsg:   "DelegatorUnbondingBalance cannot be negative",
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
