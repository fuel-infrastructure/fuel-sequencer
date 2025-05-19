package types

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestMsgUpdateParams_ValidateBasic(t *testing.T) {
	// Create a test address
	pk := ed25519.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pk.Address())

	// Use a fixed time in the future
	now := func() *time.Time {
		t := time.Now()
		return &t
	}
	futureTime := time.Now().Add(time.Hour)
	zeroTime := time.Time{}

	testCases := []struct {
		name    string
		msg     MsgUpdateParams
		wantErr bool
	}{
		{
			name: "valid message",
			msg: MsgUpdateParams{
				Authority: addr.String(),
				Params: Params{
					YieldRecipient: addr.String(),
					YieldTime:      &futureTime,
					YieldAmount:    sdkmath.NewInt(1000),
				},
			},
			wantErr: false,
		},
		{
			name: "invalid authority address",
			msg: MsgUpdateParams{
				Authority: "invalid",
				Params: Params{
					YieldRecipient: addr.String(),
					YieldTime:      &futureTime,
					YieldAmount:    sdkmath.NewInt(1000),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid yield recipient",
			msg: MsgUpdateParams{
				Authority: addr.String(),
				Params: Params{
					YieldRecipient: "invalid",
					YieldTime:      &futureTime,
					YieldAmount:    sdkmath.NewInt(1000),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid yield time (past)",
			msg: MsgUpdateParams{
				Authority: addr.String(),
				Params: Params{
					YieldRecipient: addr.String(),
					YieldTime:      &zeroTime,
					YieldAmount:    sdkmath.NewInt(1000),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid yield time (present - too close)",
			msg: MsgUpdateParams{
				Authority: addr.String(),
				Params: Params{
					YieldRecipient: addr.String(),
					YieldTime:      now(),
					YieldAmount:    sdkmath.NewInt(1000),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid yield amount (negative)",
			msg: MsgUpdateParams{
				Authority: addr.String(),
				Params: Params{
					YieldRecipient: addr.String(),
					YieldTime:      &futureTime,
					YieldAmount:    sdkmath.NewInt(-1000),
				},
			},
			wantErr: true,
		},
		{
			name: "empty params",
			msg: MsgUpdateParams{
				Authority: addr.String(),
				Params:    DefaultParams(),
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
