package types_test

import (
	"testing"

	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"
	"github.com/stretchr/testify/require"
)

func TestParams_Validate(t *testing.T) {
	validMaxSlashReportAgeBlocks := uint64(5)

	cases := []struct {
		name      string
		param     types.Params
		expectErr bool
	}{
		{
			"Default params valid",
			types.DefaultParams(),
			false,
		},
		{
			"Valid params valid",
			types.Params{
				MaxSlashReportAgeBlocks: validMaxSlashReportAgeBlocks,
			},
			false,
		},
		{
			"Invalid params - zero for MaxSlashReportAgeBlocks",
			types.Params{
				MaxSlashReportAgeBlocks: 0,
			},
			true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.param.Validate()
			if c.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
