package types

// DONTCOVER

import (
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	"github.com/ethereum/go-ethereum/common"
)

// Registered errors
var (
	ErrInvalidSigner            = sdkerrors.Register(ModuleName, 1100, "invalid signer")
	ErrInvalidSupplyDeltaPeriod = sdkerrors.Register(ModuleName, 1101, "invalid param SupplyDeltaPeriod")
	ErrUnexpectedOperation      = sdkerrors.Register(ModuleName, 1102, "operation was not expected")
	ErrInvalidSupplyDeltaValue  = sdkerrors.Register(ModuleName, 1103, "supply delta value is invalid")
	ErrInvalidEthAddress        = sdkerrors.Register(ModuleName, 1104, "invalid ethereum address")
	ErrInvalidVestingDuration   = sdkerrors.Register(ModuleName, 1105, "invalid vesting duration")
	ErrInvalidEthAddressLength  = sdkerrors.Register(ModuleName, 1106,
		fmt.Sprintf("expected eth address to be %d bytes long", common.AddressLength),
	)
	ErrCodecIsNotSupported              = sdkerrors.Register(ModuleName, 1107, "codec is not supported")
	ErrCouldNotGenerateSequencerAddress = sdkerrors.Register(
		ModuleName, 1108, "could not generate Sequencer address from Ethereum address",
	)
	ErrMsgNotAuthorizedOnSequencer = sdkerrors.Register(ModuleName, 1109, "message not authorized on Sequencer")
	ErrFailedToObtainMsgSigners    = sdkerrors.Register(ModuleName, 1110, "failed to obtain message signers")
	ErrInvalidMsgHandlerRoute      = sdkerrors.Register(ModuleName, 1111, "invalid MsgHandler route")
	ErrNilMsgResponse              = sdkerrors.Register(ModuleName, 1112, "got nil msg response")
)

// Some constant error strings used throughout the module
var (
	ErrStrOnlyProtoCodecAllowed = "only the ProtoCodec may be used for receiving messages on the Sequencer"
)
