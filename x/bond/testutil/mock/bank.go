package mock

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/mock"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

type MockBankKeeper struct {
	mock.Mock
	types.BankKeeper
}

func NewMockBankKeeper(t testing.TB) *MockBankKeeper {
	return &MockBankKeeper{}
}

// Ensure MockBankKeeper implements types.BankKeeper
var _ types.BankKeeper = (*MockBankKeeper)(nil)

func (k *MockBankKeeper) SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins {
	args := k.Called(ctx, addr)
	return args.Get(0).(sdk.Coins)
}

func (k *MockBankKeeper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	args := k.Called(ctx, senderAddr, recipientModule, amt)
	return args.Error(0)
}

func (k *MockBankKeeper) BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
	args := k.Called(ctx, moduleName, amt)
	return args.Error(0)
}

func (k *MockBankKeeper) SendCoins(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error {
	args := k.Called(ctx, fromAddr, toAddr, amt)
	return args.Error(0)
}

func (k *MockBankKeeper) GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins {
	args := k.Called(ctx, addr)
	return args.Get(0).(sdk.Coins)
}

func (k *MockBankKeeper) GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	args := k.Called(ctx, addr, denom)
	return args.Get(0).(sdk.Coin)
}

func (k *MockBankKeeper) MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
	args := k.Called(ctx, moduleName, amt)
	return args.Error(0)
}

func (k *MockBankKeeper) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	args := k.Called(ctx, senderModule, recipientAddr, amt)
	return args.Error(0)
}

func (k *MockBankKeeper) SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error {
	args := k.Called(ctx, senderModule, recipientModule, amt)
	return args.Error(0)
}
