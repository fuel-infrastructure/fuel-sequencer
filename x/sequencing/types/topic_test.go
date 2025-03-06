package types_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/fuel-infrastructure/fuel-sequencer/testutil"
	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"
	"github.com/stretchr/testify/require"
)

func TestValidateBasic(t *testing.T) {
	validAddress := "cosmos1c4k24jzduc365kywrsvf5ujz4ya6mwymy8vq4q"
	invalidAddress := "invalidAddress"

	tests := []struct {
		name   string
		topic  types.Topic
		expErr bool
	}{
		{
			name:   "valid topic",
			topic:  types.Topic{Owner: validAddress, Id: testutil.MockTopicIDHex(1), Order: math.NewInt(1)},
			expErr: false,
		},
		{
			name:   "invalid address",
			topic:  types.Topic{Owner: invalidAddress, Id: testutil.MockTopicIDHex(1), Order: math.NewInt(1)},
			expErr: true,
		},
		{
			name:   "invalid topic id",
			topic:  types.Topic{Owner: validAddress, Id: []byte{}, Order: math.NewInt(1)},
			expErr: true,
		},
		{
			name:   "invalid topic order",
			topic:  types.Topic{Owner: validAddress, Id: testutil.MockTopicIDHex(1), Order: math.NewInt(-1)},
			expErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.topic.ValidateBasic()
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateTopicId(t *testing.T) {
	tests := []struct {
		name    string
		id      interface{}
		wantErr bool
	}{
		{
			name:    "valid id",
			id:      testutil.MockTopicIDHex(1),
			wantErr: false,
		},
		{
			name:    "invalid type",
			id:      "notAnInt",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateTopicId(tc.id)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateTopicOrder(t *testing.T) {
	tests := []struct {
		name    string
		order   interface{}
		wantErr bool
	}{
		{
			name:    "valid order",
			order:   math.NewInt(1),
			wantErr: false,
		},
		{
			name:    "invalid type",
			order:   "1) What",
			wantErr: true,
		},
		{
			name:    "negative order",
			order:   math.NewInt(-1),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateTopicOrder(tc.order)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
