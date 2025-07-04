package types

import (
	"fmt"

	"cosmossdk.io/errors"
	"cosmossdk.io/math"

	"github.com/fuel-infrastructure/fuel-sequencer/utils"
)

func (t *Topic) ValidateBasic() error {

	if err := utils.ValidateBech32Address(t.Owner); err != nil {
		return errors.Wrap(err, "invalid topic owner address")
	}

	if err := ValidateTopicId(t.Id); err != nil {
		return errors.Wrap(err, "invalid topic id")
	}

	if err := ValidateTopicOrder(t.Order); err != nil {
		return errors.Wrap(err, "invalid topic order")
	}

	return nil
}

func ValidateTopicId(i interface{}) error {
	v, ok := i.([]byte)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if len(v) != 32 {
		return fmt.Errorf("expected Topic Id to be 32 bytes, received %d bytes", len(v))
	}
	return nil
}

func ValidateTopicOrder(i interface{}) error {
	v, ok := i.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.LT(math.ZeroInt()) {
		return fmt.Errorf("expected Topic Order > 0, received %d", v)
	}
	return nil
}
