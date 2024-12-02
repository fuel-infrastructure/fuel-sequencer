package types

import (
	"fmt"

	"cosmossdk.io/errors"
	"cosmossdk.io/math"
	"github.com/fuel-infrastructure/fuel-sequencer/utils"
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

	// Make sure that the value is of correct type
	v, ok := i.([]SlashEntry)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// We cannot have a slash report with no entries as we would be wasting memory
	if len(v) == 0 {
		return fmt.Errorf("entries cannot be empty")
	}

	uniqueSlashEntries := make(map[string]bool)
	for _, slashEntry := range v {

		if err := ValidateSlashEntryValidatorAddress(slashEntry.ValidatorAddress); err != nil {
			return errors.Wrap(err, "invalid SlashEntry validator address")
		}

		if err := ValidateSlashEntryDelegatorAddress(slashEntry.DelegatorAddress); err != nil {
			return errors.Wrap(err, "invalid SlashEntry delegator address")
		}

		// Make sure that there is only one entry in the SlashReport with the same (ValidatorAddress, DelegatorAddress)
		// pair. This is important for when slash reports become provable with FuelStreamX.
		slashEntryKey := fmt.Sprintf("%s|%s", slashEntry.ValidatorAddress, slashEntry.DelegatorAddress)
		if uniqueSlashEntries[slashEntryKey] {
			return ErrSlashEntryNotUnique.Wrapf("%s", slashEntryKey)
		}
		uniqueSlashEntries[slashEntryKey] = true

		if err := ValidateSlashEntryDelegatorSlashAmount(slashEntry.DelegatorSlashAmount); err != nil {
			return errors.Wrap(err, "invalid SlashEntry delegator slash amount")
		}

		if err := ValidateSlashEntryDelegatorBondedBalance(slashEntry.DelegatorBondedBalance); err != nil {
			return errors.Wrap(err, "invalid SlashEntry delegator bonded balance")
		}

		if err := ValidateSlashEntryDelegatorUnbondingBalance(slashEntry.DelegatorUnbondingBalance); err != nil {
			return errors.Wrap(err, "invalid SlashEntry delegator unbonding balance")
		}
	}

	return nil
}

func ValidateSlashEntryValidatorAddress(i interface{}) error {

	// Make sure that the value is of correct type
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// ValidatorAddress must be an operator address
	if err := utils.ValidateBech32ValAddress(v); err != nil {
		return errors.Wrapf(err, "ValidatorAddress is not a valid Bech32 Operator Address (%s)", v)
	}

	return nil
}

func ValidateSlashEntryDelegatorAddress(i interface{}) error {

	// Make sure that the value is of correct type
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// DelegatorAddress must be a valid Bech32 address
	if err := utils.ValidateBech32Address(v); err != nil {
		return errors.Wrapf(err, "DelegatorAddress is not a valid Bech32 Account address (%s)", v)
	}

	return nil
}

func ValidateSlashEntryDelegatorSlashAmount(i interface{}) error {

	// Make sure that the value is of correct type
	v, ok := i.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// DelegatorSlashAmount must be positive. We will not allow slash entries to be written in state if the slashed
	// amount is zero. We opt for this design to keep the state as pruned as possible.
	if !v.IsPositive() {
		return fmt.Errorf("DelegatorSlashAmount is not positive (%s)", v.String())
	}

	return nil
}

func ValidateSlashEntryDelegatorBondedBalance(i interface{}) error {

	// Make sure that the value is of correct type
	v, ok := i.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// Make sure that DelegatorBondedBalance is not negative
	if v.IsNegative() {
		return fmt.Errorf("DelegatorBondedBalance cannot be negative (%s)", v.String())
	}

	return nil
}

func ValidateSlashEntryDelegatorUnbondingBalance(i interface{}) error {

	// Make sure that the value is of correct type
	v, ok := i.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// Make sure that DelegatorUnbondingBalance is not negative
	if v.IsNegative() {
		return fmt.Errorf("DelegatorUnbondingBalance cannot be negative (%s)", v.String())
	}

	return nil
}
