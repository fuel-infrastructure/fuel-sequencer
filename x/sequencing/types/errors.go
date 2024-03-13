package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/sequencing module sentinel errors
var (
	ErrInvalidSigner      = sdkerrors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrSenderNotOwner     = sdkerrors.Register(ModuleName, 1101, "message sender not owner")
	ErrOrderNotMatching   = sdkerrors.Register(ModuleName, 1102, "order doesn't match")
	ErrDataTooBig         = sdkerrors.Register(ModuleName, 1103, "data from message too big")
	ErrTopicIdNotMatching = sdkerrors.Register(ModuleName, 1104, "topic id doesn't match")
	ErrInvalidGenesis     = sdkerrors.Register(ModuleName, 1105, "invalid genesis")
	ErrTopicNotUnique     = sdkerrors.Register(ModuleName, 1106, "duplicate topic found")
)
