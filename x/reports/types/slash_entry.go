package types

import (
	"fmt"

	"cosmossdk.io/errors"
	"cosmossdk.io/math"
	"github.com/fuel-infrastructure/fuel-sequencer/utils"
)

func (se *SlashEntry) ValidateBasic() error {

	if err := ValidateSlashEntryValidatorAddress(se.ValidatorAddress); err != nil {
		return errors.Wrap(err, "invalid SlashEntry validator address")
	}

	if err := ValidateSlashEntryDelegatorAddress(se.DelegatorAddress); err != nil {
		return errors.Wrap(err, "invalid SlashEntry delegator address")
	}

	if err := ValidateSlashEntryDelegatorSlashAmount(se.DelegatorSlashAmount); err != nil {
		return errors.Wrap(err, "invalid SlashEntry delegator slash amount")
	}

	if err := ValidateSlashEntryDelegatorBondedBalance(se.DelegatorBondedBalance); err != nil {
		return errors.Wrap(err, "invalid SlashEntry delegator bonded balance")
	}

	if err := ValidateSlashEntryDelegatorUnbondingBalance(se.DelegatorUnbondingBalance); err != nil {
		return errors.Wrap(err, "invalid SlashEntry delegator unbonding balance")
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
		return errors.Wrapf(err, "ValidatorAddress is not a valid Bech32 operator address (%s)", v)
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
		return errors.Wrapf(err, "DelegatorAddress is not a valid Bech32 account address (%s)", v)
	}

	return nil
}

func ValidateSlashEntryDelegatorSlashAmount(i interface{}) error {

	// Make sure that the value is of correct type
	v, ok := i.(math.LegacyDec)
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
	v, ok := i.(math.LegacyDec)
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
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// Make sure that DelegatorUnbondingBalance is not negative
	if v.IsNegative() {
		return fmt.Errorf("DelegatorUnbondingBalance cannot be negative (%s)", v.String())
	}

	return nil
}
