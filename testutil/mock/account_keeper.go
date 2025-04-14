package mock

import (
	"testing"

	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/stretchr/testify/mock"
)

type MockAccountKeeper struct {
	mock.Mock
	authkeeper.AccountKeeperI
}

func NewMockAccountKeeper(t testing.TB) *MockAccountKeeper {
	return &MockAccountKeeper{}
}

// Ensure MockAccountKeeper implements authkeeper.AccountKeeperI
var _ authkeeper.AccountKeeperI = (*MockAccountKeeper)(nil)
