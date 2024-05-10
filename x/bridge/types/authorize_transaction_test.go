package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"github.com/stretchr/testify/require"
)

func TestDeserializeAuthorizeTx(t *testing.T) {
	// Create a test codec
	registry := codectypes.NewInterfaceRegistry()
	registry.RegisterImplementations((*sdk.Msg)(nil), &banktypes.MsgSend{})
	testCodec := codec.NewProtoCodec(registry)

	// Some dummy values
	amt := sdkmath.NewInt(10)

	testCases := []struct {
		name      string
		event     *sidecartypes.AuthorizeEvent
		expMsgs   []sdk.Msg
		expErrMsg string
	}{
		{
			name:  "successfully deserialize AuthorizeTx",
			event: testtypes.TestAuthorizeEvent1,
			expMsgs: []sdk.Msg{
				&banktypes.MsgSend{
					FromAddress: testtypes.TestFrom1,
					ToAddress:   testtypes.TestTo3,
					Amount: []sdk.Coin{
						{Denom: "ufuel", Amount: amt},
					},
				},
			},
		},
		{
			name: "errors if AuthorizeTx cannot be deserialized",
			event: &sidecartypes.AuthorizeEvent{
				Sender: testtypes.TestFrom1,
				Data:   []byte(testtypes.TestData1), // Fails because the msg should be encoded to bytes using proto
			},
			expErrMsg: "proto: illegal wireType",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msgs, err := types.DeserializeAuthorizeTx(testCodec, tc.event)

			if len(tc.expErrMsg) > 0 {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expErrMsg)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expMsgs, msgs)
		})
	}
}
