// Inspired by https://github.com/cosmos/ibc-go/blob/v8.0.1/modules/apps/27-interchain-accounts/types/account.go

package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EthOwnedAccountI wraps the sdk.AccountI interface
type EthOwnedAccountI interface {
	sdk.AccountI

	// AddVestingCoins adds new coins with the specified vesting start and end times.
	// If the account is already a vesting account, the start and end time should be ignored.
	AddVestingCoins(coins sdk.Coins, startTime, endTime time.Time) (EthOwnedAccountI, error)
}

// ethOwnedAccountPretty defines an unexported struct used for encoding the EthOwnedAccount details
type ethOwnedAccountPretty struct {
	Address       sdk.AccAddress `json:"address" yaml:"address"`
	PubKey        string         `json:"public_key" yaml:"public_key"`
	AccountNumber uint64         `json:"account_number" yaml:"account_number"`
	Sequence      uint64         `json:"sequence" yaml:"sequence"`
	AccountOwner  string         `json:"account_owner" yaml:"account_owner"`
}
