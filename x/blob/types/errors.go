package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/blob module sentinel errors
var (
	ErrInvalidSigner = sdkerrors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrBlobNotFound  = sdkerrors.Register(ModuleName, 1101, "blob not found in blobpool")
)
