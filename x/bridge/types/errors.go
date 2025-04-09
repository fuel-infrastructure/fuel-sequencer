package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// Registered errors
// TODO: Some errors here need to be moved in the app folder as they are only used within the proposal handlers, not the
//
//	bridge module
var (
	ErrInvalidSigner                    = sdkerrors.Register(ModuleName, 1100, "invalid signer")
	ErrUnexpectedOperation              = sdkerrors.Register(ModuleName, 1101, "operation was not expected")
	ErrInvalidEthAddress                = sdkerrors.Register(ModuleName, 1102, "invalid ethereum address")
	ErrInvalidVestingDuration           = sdkerrors.Register(ModuleName, 1103, "invalid vesting duration")
	ErrCodecIsNotSupported              = sdkerrors.Register(ModuleName, 1104, "codec is not supported")
	ErrCouldNotGenerateSequencerAddress = sdkerrors.Register(
		ModuleName, 1105, "could not generate Sequencer address from Ethereum address",
	)
	ErrFailedToObtainMsgSigners = sdkerrors.Register(ModuleName, 1106, "failed to obtain message signers")
	ErrUnsupported              = sdkerrors.Register(ModuleName, 1107, "unsupported")
	ErrInvalidAccountAddress    = sdkerrors.Register(ModuleName, 1108, "invalid account address")
	ErrParamsInvalid            = sdkerrors.Register(ModuleName, 1109, "params are invalid")
)

// Some constant error strings used throughout the module
var (
	ErrStrOnlyProtoCodecAllowed = "only the ProtoCodec may be used for receiving messages on the Sequencer"
)
