package utils

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func ValidateBech32Address(address string) error {
	_, err := sdk.AccAddressFromBech32(address)
	return err
}
