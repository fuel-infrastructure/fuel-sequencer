package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// Registered errors
var (
	ErrInvalidSigner        = sdkerrors.Register(ModuleName, 1100, "invalid signer")
	ErrSlashReportNotUnique = sdkerrors.Register(ModuleName, 1101, "duplicate slash report found")
)
