package types_test

import (
	"testing"

	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"
	"github.com/stretchr/testify/require"
)

func TestValidateMaxBlobSizeBytes(t *testing.T) {
	cases := []struct {
		name      string
		input     interface{}
		expectErr bool
	}{
		{"Valid input", uint64(1000), false},
		{"Zero value", uint64(0), true},
		{"Invalid type", "invalid", true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := types.ValidateMaxBlobSizeBytes(c.input)
			if c.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParams_Validate(t *testing.T) {
	cases := []struct {
		name      string
		param     types.Params
		expectErr bool
	}{
		{"Valid Params", types.Params{MaxBlobSizeBytes: 1000}, false},
		{"Invalid Params", types.Params{MaxBlobSizeBytes: 0}, true},
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
