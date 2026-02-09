package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgSubmitDAAttestation{}

func (msg *MsgSubmitDAAttestation) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return ErrInvalidSigner.Wrapf("invalid sender address: %s", err)
	}

	if len(msg.BlobKey) != 32 {
		return ErrInvalidAttestation.Wrapf("blob_key must be 32 bytes, got %d", len(msg.BlobKey))
	}

	if len(msg.Attestations) == 0 {
		return ErrInvalidAttestation.Wrap("attestations must not be empty")
	}

	for i, att := range msg.Attestations {
		if len(att.ChunkHash) != 32 {
			return ErrInvalidAttestation.Wrapf("attestation[%d]: chunk_hash must be 32 bytes, got %d", i, len(att.ChunkHash))
		}
		if len(att.Signature) != 64 {
			return ErrInvalidAttestation.Wrapf("attestation[%d]: signature must be 64 bytes, got %d", i, len(att.Signature))
		}
		if att.ValidatorAddress == "" {
			return ErrInvalidAttestation.Wrapf("attestation[%d]: validator_address must not be empty", i)
		}
	}

	return nil
}

func (msg *MsgSubmitDAAttestation) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}
