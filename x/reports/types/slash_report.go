package types

import (
	"fmt"

	"cosmossdk.io/errors"
)

func (sr *SlashReport) ValidateBasic() error {

	if err := ValidateSlashReportHeight(sr.Height); err != nil {
		return errors.Wrap(err, "invalid SlashReport height")
	}

	if err := ValidateSlashReportEntries(sr.Entries); err != nil {
		return errors.Wrap(err, "invalid SlashReport entries")
	}

	return nil
}

func ValidateSlashReportHeight(i interface{}) error {

	// Make sure that the value is of correct type
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// Height must be bigger than zero because validators can't be slashed on genesis. Being uint64, negative heights
	// are automatically omitted.
	if v == 0 {
		return fmt.Errorf("slash report height cannot be zero")
	}

	return nil
}

func ValidateSlashReportEntries(i interface{}) error {

	// Make sure that the value is of correct type.
	v, ok := i.([]SlashEntry)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// We cannot have a slash report with no entries as we would be wasting memory.
	if len(v) == 0 {
		return fmt.Errorf("entries cannot be empty")
	}

	uniqueSlashEntries := make(map[string]bool)
	for _, slashEntry := range v {

		// Make sure that the slash entry satisfies the basic validation rules.
		if err := slashEntry.ValidateBasic(); err != nil {
			return err
		}

		// Make sure that there is only one entry in the SlashReport with the same (ValidatorAddress, DelegatorAddress)
		// pair. Even though the slash report height is used to construct the storage keys, it is not required at this
		// stage because all slash entries within the same slash report are assigned the same height.
		slashEntryKey := fmt.Sprintf("%s|%s", slashEntry.DelegatorAddress, slashEntry.ValidatorAddress)
		if uniqueSlashEntries[slashEntryKey] {
			return ErrSlashEntryNotUnique.Wrapf("slash entry not unique at key %s", slashEntryKey)
		}
		uniqueSlashEntries[slashEntryKey] = true
	}

	return nil
}
