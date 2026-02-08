package types

import (
	"fmt"
	"io"
)

// Marshal/Unmarshal/Size methods for DAChunkSig, MsgSubmitDAAttestation,
// and MsgSubmitDAAttestationResponse. These follow the gogoproto pattern
// used by other types in tx.pb.go.

// --- DAChunkSig ---

func (m *DAChunkSig) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *DAChunkSig) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *DAChunkSig) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	// field 4: signature (bytes)
	if len(m.Signature) > 0 {
		i -= len(m.Signature)
		copy(dAtA[i:], m.Signature)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Signature)))
		i--
		dAtA[i] = 0x22
	}
	// field 3: chunk_hash (bytes)
	if len(m.ChunkHash) > 0 {
		i -= len(m.ChunkHash)
		copy(dAtA[i:], m.ChunkHash)
		i = encodeVarintTx(dAtA, i, uint64(len(m.ChunkHash)))
		i--
		dAtA[i] = 0x1a
	}
	// field 2: chunk_index (uint32)
	if m.ChunkIndex != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.ChunkIndex))
		i--
		dAtA[i] = 0x10
	}
	// field 1: validator_address (string)
	if len(m.ValidatorAddress) > 0 {
		i -= len(m.ValidatorAddress)
		copy(dAtA[i:], m.ValidatorAddress)
		i = encodeVarintTx(dAtA, i, uint64(len(m.ValidatorAddress)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *DAChunkSig) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ValidatorAddress)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.ChunkIndex != 0 {
		n += 1 + sovTx(uint64(m.ChunkIndex))
	}
	l = len(m.ChunkHash)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.Signature)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *DAChunkSig) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: DAChunkSig: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: DAChunkSig: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1: // validator_address
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ValidatorAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2: // chunk_index
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ChunkIndex", wireType)
			}
			m.ChunkIndex = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ChunkIndex |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3: // chunk_hash
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ChunkHash", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ChunkHash = append(m.ChunkHash[:0], dAtA[iNdEx:postIndex]...)
			if m.ChunkHash == nil {
				m.ChunkHash = []byte{}
			}
			iNdEx = postIndex
		case 4: // signature
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Signature", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Signature = append(m.Signature[:0], dAtA[iNdEx:postIndex]...)
			if m.Signature == nil {
				m.Signature = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if iNdEx+skippy > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}

// --- MsgSubmitDAAttestation ---

func (m *MsgSubmitDAAttestation) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgSubmitDAAttestation) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgSubmitDAAttestation) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	// field 4: attestations (repeated message)
	if len(m.Attestations) > 0 {
		for iNdEx := len(m.Attestations) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Attestations[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintTx(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x22
		}
	}
	// field 3: blob_size (uint64)
	if m.BlobSize != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.BlobSize))
		i--
		dAtA[i] = 0x18
	}
	// field 2: blob_key (bytes)
	if len(m.BlobKey) > 0 {
		i -= len(m.BlobKey)
		copy(dAtA[i:], m.BlobKey)
		i = encodeVarintTx(dAtA, i, uint64(len(m.BlobKey)))
		i--
		dAtA[i] = 0x12
	}
	// field 1: sender (string)
	if len(m.Sender) > 0 {
		i -= len(m.Sender)
		copy(dAtA[i:], m.Sender)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Sender)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgSubmitDAAttestation) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Sender)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.BlobKey)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.BlobSize != 0 {
		n += 1 + sovTx(uint64(m.BlobSize))
	}
	if len(m.Attestations) > 0 {
		for _, e := range m.Attestations {
			l = e.Size()
			n += 1 + l + sovTx(uint64(l))
		}
	}
	return n
}

func (m *MsgSubmitDAAttestation) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgSubmitDAAttestation: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgSubmitDAAttestation: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1: // sender
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Sender", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Sender = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2: // blob_key
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BlobKey", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.BlobKey = append(m.BlobKey[:0], dAtA[iNdEx:postIndex]...)
			if m.BlobKey == nil {
				m.BlobKey = []byte{}
			}
			iNdEx = postIndex
		case 3: // blob_size
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field BlobSize", wireType)
			}
			m.BlobSize = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.BlobSize |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4: // attestations
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Attestations", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Attestations = append(m.Attestations, DAChunkSig{})
			if err := m.Attestations[len(m.Attestations)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if iNdEx+skippy > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}

// --- MsgSubmitDAAttestationResponse ---

func (m *MsgSubmitDAAttestationResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgSubmitDAAttestationResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgSubmitDAAttestationResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	// field 3: attested_power (string)
	if len(m.AttestedPower) > 0 {
		i -= len(m.AttestedPower)
		copy(dAtA[i:], m.AttestedPower)
		i = encodeVarintTx(dAtA, i, uint64(len(m.AttestedPower)))
		i--
		dAtA[i] = 0x1a
	}
	// field 2: total_power (string)
	if len(m.TotalPower) > 0 {
		i -= len(m.TotalPower)
		copy(dAtA[i:], m.TotalPower)
		i = encodeVarintTx(dAtA, i, uint64(len(m.TotalPower)))
		i--
		dAtA[i] = 0x12
	}
	// field 1: confirmed (bool)
	if m.Confirmed {
		i--
		if m.Confirmed {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *MsgSubmitDAAttestationResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Confirmed {
		n += 1 + 1
	}
	l = len(m.TotalPower)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.AttestedPower)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgSubmitDAAttestationResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgSubmitDAAttestationResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgSubmitDAAttestationResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1: // confirmed
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Confirmed", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Confirmed = bool(v != 0)
		case 2: // total_power
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TotalPower", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.TotalPower = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3: // attested_power
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AttestedPower", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AttestedPower = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if iNdEx+skippy > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
