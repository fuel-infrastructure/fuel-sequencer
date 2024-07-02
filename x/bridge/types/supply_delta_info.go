package types

import (
	"cosmossdk.io/errors"
	"cosmossdk.io/math"
	"fmt"
)

func (t *SupplyDeltaInfo) ValidateBasic() error {

	if err := ValidateLastSupply(t.LastSupply); err != nil {
		return errors.Wrap(err, "invalid LastSupply")
	}

	if err := ValidateDelta(t.Delta); err != nil {
		return errors.Wrap(err, "invalid Delta")
	}

	if err := ValidateOffset(t.Offset); err != nil {
		return errors.Wrap(err, "invalid Offset")
	}

	return nil
}

func ValidateLastSupply(i interface{}) error {
	v, ok := i.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("expected LastSupply >= 0, received %s", v.String())
	}
	return nil
}

func ValidateDelta(i interface{}) error {
	_, ok := i.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func ValidateOffset(i interface{}) error {
	_, ok := i.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}
