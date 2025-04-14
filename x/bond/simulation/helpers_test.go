package simulation

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/stretchr/testify/require"
)

func TestFindAccount(t *testing.T) {
	// Create test accounts
	acc1 := simtypes.Account{
		Address: sdk.AccAddress("test1"),
		PrivKey: nil,
		PubKey:  nil,
	}
	acc2 := simtypes.Account{
		Address: sdk.AccAddress("test2"),
		PrivKey: nil,
		PubKey:  nil,
	}
	accs := []simtypes.Account{acc1, acc2}

	tests := []struct {
		name      string
		address   string
		wantAcc   simtypes.Account
		wantFound bool
		wantPanic bool
	}{
		{
			name:      "find existing account",
			address:   sdk.AccAddress("test1").String(),
			wantAcc:   acc1,
			wantFound: true,
			wantPanic: false,
		},
		{
			name:      "find non-existent account",
			address:   sdk.AccAddress("test3").String(),
			wantAcc:   simtypes.Account{},
			wantFound: false,
			wantPanic: false,
		},
		{
			name:      "invalid bech32 address",
			address:   "invalid",
			wantAcc:   simtypes.Account{},
			wantFound: false,
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				require.Panics(t, func() {
					FindAccount(accs, tt.address)
				})
				return
			}

			gotAcc, gotFound := FindAccount(accs, tt.address)
			require.Equal(t, tt.wantFound, gotFound)
			if tt.wantFound {
				require.Equal(t, tt.wantAcc.Address, gotAcc.Address)
			}
		})
	}
}
