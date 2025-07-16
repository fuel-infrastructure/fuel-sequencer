package types

// DONTCOVER

import (
	"cosmossdk.io/errors"
)

// x/blob module sentinel errors
var (
	ErrBlobNotFound  = errors.Register(ModuleName, 1, "blob not found")
	ErrInvalidSigner = errors.Register(ModuleName, 2, "invalid signer")
)
