package types

import (
	"encoding/json"
	"fmt"

	proto "github.com/cosmos/gogoproto/proto"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DAChunkSig represents a single chunk attestation signature from a validator.
type DAChunkSig struct {
	ValidatorAddress string `protobuf:"bytes,1,opt,name=validator_address,json=validatorAddress,proto3" json:"validator_address,omitempty"`
	ChunkIndex       uint32 `protobuf:"varint,2,opt,name=chunk_index,json=chunkIndex,proto3" json:"chunk_index,omitempty"`
	ChunkHash        []byte `protobuf:"bytes,3,opt,name=chunk_hash,json=chunkHash,proto3" json:"chunk_hash,omitempty"`
	Signature        []byte `protobuf:"bytes,4,opt,name=signature,proto3" json:"signature,omitempty"`
}

func (m *DAChunkSig) Reset()         { *m = DAChunkSig{} }
func (m *DAChunkSig) String() string { return proto.CompactTextString(m) }
func (*DAChunkSig) ProtoMessage()    {}
func (*DAChunkSig) Descriptor() ([]byte, []int) {
	return fileDescriptor_55559594e2513708, []int{4}
}

var xxx_messageInfo_DAChunkSig proto.InternalMessageInfo

func (m *DAChunkSig) GetValidatorAddress() string {
	if m != nil {
		return m.ValidatorAddress
	}
	return ""
}
func (m *DAChunkSig) GetChunkIndex() uint32 {
	if m != nil {
		return m.ChunkIndex
	}
	return 0
}
func (m *DAChunkSig) GetChunkHash() []byte {
	if m != nil {
		return m.ChunkHash
	}
	return nil
}
func (m *DAChunkSig) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

// MsgSubmitDAAttestation submits DA attestation signatures for on-chain verification.
type MsgSubmitDAAttestation struct {
	Sender       string       `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
	BlobKey      []byte       `protobuf:"bytes,2,opt,name=blob_key,json=blobKey,proto3" json:"blob_key,omitempty"`
	BlobSize     uint64       `protobuf:"varint,3,opt,name=blob_size,json=blobSize,proto3" json:"blob_size,omitempty"`
	Attestations []DAChunkSig `protobuf:"bytes,4,rep,name=attestations,proto3" json:"attestations"`
}

func (msg *MsgSubmitDAAttestation) Reset()         { *msg = MsgSubmitDAAttestation{} }
func (msg *MsgSubmitDAAttestation) String() string { b, _ := json.Marshal(msg); return string(b) }
func (*MsgSubmitDAAttestation) ProtoMessage()      {}
func (*MsgSubmitDAAttestation) Descriptor() ([]byte, []int) {
	return fileDescriptor_55559594e2513708, []int{5}
}

var xxx_messageInfo_MsgSubmitDAAttestation proto.InternalMessageInfo

func (m *MsgSubmitDAAttestation) GetSender() string {
	if m != nil {
		return m.Sender
	}
	return ""
}
func (m *MsgSubmitDAAttestation) GetBlobKey() []byte {
	if m != nil {
		return m.BlobKey
	}
	return nil
}
func (m *MsgSubmitDAAttestation) GetBlobSize() uint64 {
	if m != nil {
		return m.BlobSize
	}
	return 0
}
func (m *MsgSubmitDAAttestation) GetAttestations() []DAChunkSig {
	if m != nil {
		return m.Attestations
	}
	return nil
}

// MsgSubmitDAAttestationResponse is the response from SubmitDAAttestation.
type MsgSubmitDAAttestationResponse struct {
	Confirmed     bool   `protobuf:"varint,1,opt,name=confirmed,proto3" json:"confirmed,omitempty"`
	TotalPower    string `protobuf:"bytes,2,opt,name=total_power,json=totalPower,proto3" json:"total_power,omitempty"`
	AttestedPower string `protobuf:"bytes,3,opt,name=attested_power,json=attestedPower,proto3" json:"attested_power,omitempty"`
}

func (msg *MsgSubmitDAAttestationResponse) Reset()         { *msg = MsgSubmitDAAttestationResponse{} }
func (msg *MsgSubmitDAAttestationResponse) String() string { return fmt.Sprintf("%+v", *msg) }
func (*MsgSubmitDAAttestationResponse) ProtoMessage()      {}
func (*MsgSubmitDAAttestationResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_55559594e2513708, []int{6}
}

var xxx_messageInfo_MsgSubmitDAAttestationResponse proto.InternalMessageInfo

func (m *MsgSubmitDAAttestationResponse) GetConfirmed() bool {
	if m != nil {
		return m.Confirmed
	}
	return false
}
func (m *MsgSubmitDAAttestationResponse) GetTotalPower() string {
	if m != nil {
		return m.TotalPower
	}
	return ""
}
func (m *MsgSubmitDAAttestationResponse) GetAttestedPower() string {
	if m != nil {
		return m.AttestedPower
	}
	return ""
}

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
