package sample

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccAddress returns a sample account address
func AccAddress() string {
	return AccAddressBz().String()
}

// AccAddressBz returns a sample account address without converting to string
func AccAddressBz() sdk.AccAddress {
	pk := ed25519.GenPrivKey().PubKey()
	addr := pk.Address()
	return sdk.AccAddress(addr)
}

// ValAddress returns a sample validator address
func ValAddress() string {
	return ValAddressBz().String()
}

// ValAddressBz returns a sample validator address without converting to string
func ValAddressBz() sdk.ValAddress {
	pk := ed25519.GenPrivKey().PubKey()
	addr := pk.Address()
	return sdk.ValAddress(addr)
}
