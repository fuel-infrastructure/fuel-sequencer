package types

import (
	comettypes "github.com/cometbft/cometbft/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var _ paramtypes.ParamSet = (*Params)(nil)

const (
	DefaultMaxBlobSize = 10 * 1024 * 1024 // 10 MB

	// DefaultSequencerTxMaxBytes is the default max size in bytes for Sequencer-native transactions. This is set
	// to 20971520 bytes assuming a max block size of 22020096 bytes (21 MB, Comet BFT default) and a large buffer to
	// reserve block space for non-tx data (header, evidence etc.) and MsgIndexTx size.
	DefaultSequencerTxMaxBytes = 20971520 // 20 MB
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

	// If the MaxBlobSizeBytes is greater than the cometBFT MaxBlobSizeBytes we reject it.
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
