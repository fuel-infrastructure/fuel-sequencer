package types_test

import (
	"errors"
	"testing"

	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	"github.com/stretchr/testify/require"
)

func TestIsErrorFatal(t *testing.T) {
	testCases := []struct {
		name           string
		error          error
		expectedResult bool
	}{
		{
			name:           "IsErrorFatal returns false if ErrSidecarFallenBehindWithAcceptableDelay",
			error:          errors.New(types.ErrSidecarFallenBehindWithAcceptableDelay),
			expectedResult: false,
		},
		{
			name:           "IsErrorFatal returns false if ErrBlockDoesNotExist",
			error:          errors.New(types.ErrBlockDoesNotExist),
			expectedResult: false,
		},
		{
			name:           "IsErrorFatal returns true if dummy error",
			error:          errors.New("dummy error"),
			expectedResult: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualResult := types.IsErrorFatal(tc.error)
			require.Equal(t, tc.expectedResult, actualResult)
		})
	}
}
