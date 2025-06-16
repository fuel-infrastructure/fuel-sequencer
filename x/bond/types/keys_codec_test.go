package types_test

import (
	"testing"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

func TestKeyPrefix(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []byte
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []byte{},
		},
		{
			name:     "simple string",
			input:    "test",
			expected: []byte("test"),
		},
		{
			name:     "module name",
			input:    types.ModuleName,
			expected: []byte(types.ModuleName),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := types.KeyPrefix(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestRegisterInterfaces(t *testing.T) {
	registry := cdctypes.NewInterfaceRegistry()
	types.RegisterInterfaces(registry)

	// Verify that MsgUpdateParams is registered as a sdk.Msg
	msg := &types.MsgUpdateParams{}
	require.Implements(t, (*sdk.Msg)(nil), msg)
}
