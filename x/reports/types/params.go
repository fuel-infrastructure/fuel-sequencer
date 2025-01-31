package types

const (
	// DefaultMaxSlashReportAgeBlocks is 1 day assuming an average block time of 6 seconds.
	DefaultMaxSlashReportAgeBlocks = 14400
)

// NewParams creates a new Params instance
func NewParams(
	maxSlashReportAgeBlocks uint64,
) Params {
	return Params{
		MaxSlashReportAgeBlocks: maxSlashReportAgeBlocks,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		DefaultMaxSlashReportAgeBlocks,
	)
}

// Validate validates the set of params
func (p Params) Validate() error {

	if err := ValidateMaxSlashReportAgeBlocks(p.MaxSlashReportAgeBlocks); err != nil {
		return err
	}

	return nil
}

func ValidateMaxSlashReportAgeBlocks(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return ErrParamsInvalid.Wrapf("invalid parameter type: %T", i)
	}
	if v <= 0 {
		return ErrParamsInvalid.Wrapf("max slash report age blocks must be positive, got: %d", v)
	}
	return nil
}
