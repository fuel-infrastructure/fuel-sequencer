package types_test

import (
	"testing"

	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"
	"github.com/stretchr/testify/require"
)

func TestParams_Validate(t *testing.T) {
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
		// TODO: Add valid and invalid param test cases when params are implemented
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
