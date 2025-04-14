package mock

import (
	"testing"

	"cosmossdk.io/core/address"
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

// AddressCodec returns the mock's address codec
func (m *MockAccountKeeper) AddressCodec() address.Codec {
	args := m.Called()
	return args.Get(0).(address.Codec)
}
