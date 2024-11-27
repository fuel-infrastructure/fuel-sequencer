package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// Registered errors
var (
	ErrInvalidSigner = sdkerrors.Register(ModuleName, 1100, "invalid signer")
	ErrSample        = sdkerrors.Register(ModuleName, 1101, "sample error")
)
