package keeper

import (
	"context"
	"fmt"
	"testing"

	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/testutil"
	bondtypes "github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

type mockAddressCodec struct{}

func (m mockAddressCodec) StringToBytes(text string) ([]byte, error) {
	if text == "invalid-address" {
		return nil, fmt.Errorf("invalid address")
	}
	return []byte(text), nil
}

func (m mockAddressCodec) BytesToString(bz []byte) (string, error) {
	return string(bz), nil
}

func setupTest(t *testing.T) (Keeper, context.Context) {
	ctrl := gomock.NewController(t)
	cdc := codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
	storeService := testutil.NewMockStoreService(ctrl)
	logger := log.NewNopLogger()

	// Create a valid bech32 address
	addr := sdk.AccAddress([]byte("test-address"))
	authority := addr.String()

	accountKeeper := testutil.NewMockAccountKeeper(ctrl)
	accountKeeper.EXPECT().AddressCodec().Return(mockAddressCodec{}).AnyTimes()

	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// Set up store expectations
	kvStore := testutil.NewMockKVStore(ctrl)
	storeService.EXPECT().OpenKVStore(gomock.Any()).Return(kvStore).AnyTimes()

	// Set up state storage expectations
	var storedState []byte
	kvStore.EXPECT().Get(gomock.Any()).DoAndReturn(func(key []byte) ([]byte, error) {
		return storedState, nil
	}).AnyTimes()
	kvStore.EXPECT().Set(gomock.Any(), gomock.Any()).DoAndReturn(func(key, value []byte) error {
		storedState = value
		return nil
	}).AnyTimes()

	k := NewKeeper(cdc, storeService, logger, authority, accountKeeper, bankKeeper)
	ctx := context.Background()
	return k, ctx
}

func TestGetSetState(t *testing.T) {
	k, ctx := setupTest(t)

	// Test GetState with empty store
	state := k.GetState(ctx)
	require.Equal(t, bondtypes.DefaultState(), state)

	// Test SetState
	newState := bondtypes.State{
		YieldMintHeight: 100,
	}
	err := k.SetState(ctx, newState)
	require.NoError(t, err)

	// Test GetState after setting
	state = k.GetState(ctx)
	require.Equal(t, newState, state)
}

func TestGetSetYieldMintHeight(t *testing.T) {
	k, ctx := setupTest(t)

	// Test GetYieldMintHeight with default state
	height := k.GetYieldMintHeight(ctx)
	require.Equal(t, int64(0), height)

	// Test SetYieldMintHeight
	newHeight := int64(100)
	err := k.SetYieldMintHeight(ctx, newHeight)
	require.NoError(t, err)

	// Test GetYieldMintHeight after setting
	height = k.GetYieldMintHeight(ctx)
	require.Equal(t, newHeight, height)
}

func TestGetAccountAsBytes(t *testing.T) {
	k, _ := setupTest(t)

	// Test valid address
	addr := sdk.AccAddress([]byte("test-address"))
	validAddr := addr.String()
	bytes, err := k.GetAccountAsBytes(validAddr)
	require.NoError(t, err)
	require.NotNil(t, bytes)

	// Test invalid address
	invalidAddr := "invalid-address"
	bytes, err = k.GetAccountAsBytes(invalidAddr)
	require.Error(t, err)
	require.Nil(t, bytes)
}
