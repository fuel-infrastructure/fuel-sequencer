package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/bond module sentinel errors
var (
	ErrInvalidSigner = sdkerrors.Register(ModuleName, 1100, "invalid signer")
)
