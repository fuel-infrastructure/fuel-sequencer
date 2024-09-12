package types

import (
	comettypes "github.com/cometbft/cometbft/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var _ paramtypes.ParamSet = (*Params)(nil)

const (
	DefaultMaxBlobSize = 1_048_576 // 1 MiB

	// DefaultSequencerTxMaxBytes is the default max size in bytes for Sequencer-native transactions. This is set to
	// 1153434 bytes with the assumption that it will fit in the max block size configured for the Sequencer, even after
	// accounting for the block space taken up by non-tx data (header, evidence etc.) and MsgIndex transaction size.
	// Ref for 'non-tx data': https://github.com/cometbft/cometbft/blob/v0.38.6/types/block.go#L278.
	//
	// If the chain is configured with DefaultSequencerTxMaxBytes, MaxBlobSizeBytes should not exceed 1048576 bytes
	// (1MiB). Violating this condition may result in MsgPostBlob transactions being rejected.
	DefaultSequencerTxMaxBytes = 1_153_434 // ~1.1 MiB
)

// NewParams creates a new Params instance
func NewParams(maxBlobSizeBytes, sequencerTxMaxBytes uint64) Params {
	return Params{
		MaxBlobSizeBytes:    maxBlobSizeBytes,
		SequencerTxMaxBytes: sequencerTxMaxBytes,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultMaxBlobSize, DefaultSequencerTxMaxBytes)
}

// ParamSetPairs implements params.ParamSet
//
// Deprecated.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{}
}

// Validate validates the set of params
func (p Params) Validate() error {

	// Validate max blob size bytes is the maximum size of a blob the sequencer can accept.
	if err := ValidateMaxBlobSizeBytes(p.MaxBlobSizeBytes); err != nil {
		return err
	}

	// Validate sequencer transaction max bytes.
	if err := ValidateSequencerTxMaxBytes(p.SequencerTxMaxBytes); err != nil {
		return err
	}

	return nil
}

// ValidateMaxBlobSizeBytes verifies that the max blob size bytes is of the correct type and is greater than 0.
func ValidateMaxBlobSizeBytes(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return ErrParamsInvalid.Wrapf(
			"invalid parameter type for maxBlobSizeBytes: %T",
			i,
		)
	}

	// Check that MaxBlobSizeBytes is greater than 0.
	if v == 0 {
		return ErrParamsInvalid.Wrapf(
			"maxBlobSizeBytes must be greater than 0",
		)
	}

	// If the MaxBlobSizeBytes is greater than the cometBFT MaxBlockSizeBytes we reject it.
	if v > comettypes.MaxBlockSizeBytes {
		return ErrParamsInvalid.Wrapf(
			"maxBlobSizeBytes %d cannot be greater than cometbft max block size %d",
			v, comettypes.MaxBlockSizeBytes,
		)
	}

	return nil
}

// ValidateSequencerTxMaxBytes verifies that the max Sequencer transaction bytes is of the correct type and is positive.
func ValidateSequencerTxMaxBytes(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return ErrParamsInvalid.Wrapf("invalid parameter type for sequencerTxMaxBytes: %T", i)
	}

	// Check that SequencerTxMaxBytes is greater than 0.
	if v <= 0 {
		return ErrParamsInvalid.Wrapf("sequencerTxMaxBytes must be greater than 0")
	}

	// If SequencerTxMaxBytes is greater than the cometBFT MaxBlockSizeBytes we reject it.
	if v > comettypes.MaxBlockSizeBytes {
		return ErrParamsInvalid.Wrapf(
			"sequencerTxMaxBytes %d cannot be greater than cometbft max block size %d",
			v, comettypes.MaxBlockSizeBytes,
		)
	}

	return nil
}
