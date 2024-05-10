package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
)

// DeserializeAuthorizeTx attempts to unmarshal AuthorizeEvent.Data into AuthorizeTx and unpacks the Any messages
// in AuthorizeTx.Data into an array of sdk.Msg
func DeserializeAuthorizeTx(cdc codec.BinaryCodec, event *sidecartypes.AuthorizeEvent) ([]sdk.Msg, error) {
	// this is a defensive check to ensure only the ProtoCodec is used for message unmarshalling
	if _, ok := cdc.(*codec.ProtoCodec); !ok {
		return nil, ErrCodecIsNotSupported.Wrap(ErrStrOnlyProtoCodecAllowed)
	}

	var authorizeTx AuthorizeTx
	if err := cdc.Unmarshal(event.Data, &authorizeTx); err != nil {
		return nil, err
	}

	msgs := make([]sdk.Msg, len(authorizeTx.Messages))

	for i, protoAny := range authorizeTx.Messages {
		var msg sdk.Msg
		err := cdc.UnpackAny(protoAny, &msg)
		if err != nil {
			return nil, err
		}
		msgs[i] = msg
	}

	return msgs, nil
}
