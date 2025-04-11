package mock

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/mock"
)

type MockBankKeeper struct {
	mock.Mock
}

func NewMockBankKeeper(t testing.TB) *MockBankKeeper {
	return &MockBankKeeper{}
}

func (m *MockBankKeeper) SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	args := m.Called(ctx, addr)
	return args.Get(0).(sdk.Coins)
}

func (m *MockBankKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	args := m.Called(ctx, senderAddr, recipientModule, amt)
	return args.Error(0)
}

func (m *MockBankKeeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	args := m.Called(ctx, moduleName, amt)
	return args.Error(0)
}
