package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/blob module sentinel errors
var (
	ErrInvalidSigner = sdkerrors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")

	// BlobPool errors
	ErrBlobNotFound = sdkerrors.Register(ModuleName, 1101, "blob not found in blobpool")

	// MsgBlobMetadataTx errors
	ErrMsgBlobMetadataTxNil = sdkerrors.Register(ModuleName, 1111, "nil MsgBlobMetadataTx")
	ErrHashEmpty            = sdkerrors.Register(ModuleName, 1112, "metadata-specified hash is empty")
	ErrHashNotValidKey      = sdkerrors.Register(ModuleName, 1113, "metadata-specified hash is not a valid key")
	ErrSizeZero             = sdkerrors.Register(ModuleName, 1114, "metadata-specified size is 0")
	ErrTopicEmpty           = sdkerrors.Register(ModuleName, 1115, "metadata-specified topic is empty")
	ErrNonceZero            = sdkerrors.Register(ModuleName, 1116, "metadata-specified nonce is 0")
)
