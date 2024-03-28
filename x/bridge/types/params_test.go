package types_test

import (
	"testing"
	"time"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"github.com/stretchr/testify/require"
)

func TestVestingTimesFromVestingDuration(t *testing.T) {

	// Helper durations.
	years1 := time.Hour * 24 * 365
	years2 := years1 * 2
	years4 := years1 * 4

	// Helper times.
	t0, _ := time.Parse(time.DateOnly, "2024-01-01")
	t0Plus1Year := t0.Add(years1)  // accounts for vesting start time delay
	t0Plus2Years := t0.Add(years2) // used for 2-year vesting duration
	t0Plus4Years := t0.Add(years4) // used for 4-year vesting duration

	// Fixed params.
	params := types.Params{
		VestingStartTime: t0,
	}

	testCases := []struct {
		name            string
		vestingDuration time.Duration
		expStartTime    time.Time
		expEndTime      time.Time
		expErrMsg       string
	}{
		{
			name:            "Zero vesting duration is less than vestingStartTimeDelay => err",
			vestingDuration: 0,
			expErrMsg:       "must be greater than vesting start time delay, got 0s <= 8760h0m0s",
		},
		{
			name:            "Almost 1 year vesting duration is less than vestingStartTimeDelay => err",
			vestingDuration: years1 - 1,
			expErrMsg:       "must be greater than vesting start time delay, got 8759h59m59.999999999s <= 8760h0m0s",
		},
		{
			name:            "1 year vesting duration is equal to vestingStartTimeDelay => err",
			vestingDuration: years1,
			expErrMsg:       "must be greater than vesting start time delay, got 8760h0m0s <= 8760h0m0s",
		},
		{
			name:            "Just over 1 year vesting duration is valid",
			vestingDuration: years1 + 1,
			expStartTime:    t0Plus1Year,        // t0 + 1 year
			expEndTime:      t0Plus1Year.Add(1), // t0 + vestingDuration
		},
		{
			name:            "2 years => 1 year lock followed by 1 year vesting",
			vestingDuration: years2,
			expStartTime:    t0Plus1Year,  // t0 + 1 year
			expEndTime:      t0Plus2Years, // t0 + vestingDuration
		},
		{
			name:            "4 years => 1 year lock followed by 3 year vesting",
			vestingDuration: years4,
			expStartTime:    t0Plus1Year,  // t0 + 1 year
			expEndTime:      t0Plus4Years, // t0 + vestingDuration
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			startTime, endTime, err := params.VestingTimesFromVestingDuration(tc.vestingDuration)
			if tc.expErrMsg != "" {
				require.ErrorContains(t, err, tc.expErrMsg)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expStartTime, startTime)
			require.Equal(t, tc.expEndTime, endTime)
		})
	}
}
