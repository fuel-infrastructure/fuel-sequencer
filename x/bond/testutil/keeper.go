package testutil

import (
	"testing"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/testutil/mock"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

func NewMockAccountKeeper(t testing.TB) types.AccountKeeper {
	return mock.NewMockAccountKeeper(t)
}

func NewMockBankKeeper(t testing.TB) types.BankKeeper {
	return mock.NewMockBankKeeper(t)
}
