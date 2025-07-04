package testutil

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/testutil/fixtures"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func EventFromMsg(t *testing.T, msg sdk.Msg) *sidecartypes.Event {
	anyMsgs, err := sidecartypes.NewAnysWithValue(msg)
	require.NoError(t, err)

	// Serialize the message to bytes
	bz, err := proto.Marshal(&bridgetypes.AuthorizeTx{Messages: anyMsgs})
	require.NoError(t, err)

	authorizeEvent := sidecartypes.AuthorizeEvent{
		Sender: fixtures.SenderAddress,
		Data:   bz,
	}
	authorizeEventBz, err := authorizeEvent.Marshal()
	require.NoError(t, err)

	event := &sidecartypes.Event{
		EventType:       sidecartypes.AuthorizeEventName,
		Data:            authorizeEventBz,
		ContractAddress: fixtures.SequencerProxyContractAddress,
	}
	return event
}

func EventFromDepositEvent(t *testing.T, depositEvent sidecartypes.DepositEvent) *sidecartypes.Event {
	depositEventBz, err := depositEvent.Marshal()
	require.NoError(t, err)

	return &sidecartypes.Event{
		EventType:       sidecartypes.DepositEventName,
		Data:            depositEventBz,
		ContractAddress: fixtures.SequencerProxyContractAddress,
	}
}
