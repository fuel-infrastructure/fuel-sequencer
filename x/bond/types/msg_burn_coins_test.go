package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

// TestNewMsgBurnCoins tests the creation and basic validation of MsgBurnCoins.
// Note: Address validation is intentionally not performed in ValidateBasic() to allow
// for more flexible address formats (like hex addresses) to be handled during message
// execution. This approach:
// 1. Keeps the message type simple and focused on structural validation
// 2. Allows for different address validation strategies in different contexts
// 3. Enables support for multiple address formats without changing the message type
// 4. Maintains backward compatibility with existing code
func TestNewMsgBurnCoins(t *testing.T) {
	testCases := []struct {
		name   string
		sender string
		coins  sdk.Coins
		expErr bool
	}{
		{
			name:   "valid message with bech32 address",
			sender: "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
			coins:  sdk.NewCoins(sdk.NewCoin("ufuel", sdkmath.NewInt(100))),
			expErr: false,
		},
		{
			name:   "valid message with hex address",
			sender: "0x1234567890123456789012345678901234567890",
			coins:  sdk.NewCoins(sdk.NewCoin("ufuel", sdkmath.NewInt(100))),
			expErr: false,
		},
		{
			name:   "invalid sender",
			sender: "invalid",
			coins:  sdk.NewCoins(sdk.NewCoin("ufuel", sdkmath.NewInt(100))),
			expErr: false, // No validation in ValidateBasic
		},
		{
			name:   "invalid hex address",
			sender: "0xinvalid",
			coins:  sdk.NewCoins(sdk.NewCoin("ufuel", sdkmath.NewInt(100))),
			expErr: false, // No validation in ValidateBasic
		},
		{
			name:   "zero coins",
			sender: "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
			coins:  sdk.NewCoins(),
			expErr: true,
		},
		{
			name:   "negative coins",
			sender: "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
			coins:  sdk.Coins{sdk.Coin{Denom: "ufuel", Amount: sdkmath.NewInt(-100)}},
			expErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgBurnCoins(tc.sender, tc.coins)
			err := msg.ValidateBasic()
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.sender, msg.Sender)
				require.Equal(t, tc.coins, msg.Coins)
			}
		})
	}
}

// TestMsgBurnCoins_GetSigners tests the GetSigners method of MsgBurnCoins.
// Note: GetSigners is used for signature verification and must use the SDK's
// address format. Address validation is performed here to ensure proper signature
// verification.
func TestMsgBurnCoins_GetSigners(t *testing.T) {
	testCases := []struct {
		name   string
		sender string
		expErr bool
	}{
		{
			name:   "valid bech32 sender",
			sender: "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
			expErr: false,
		},
		{
			name:   "valid hex sender",
			sender: "0x1234567890123456789012345678901234567890",
			expErr: true, // GetSigners requires Bech32 format
		},
		{
			name:   "invalid sender",
			sender: "invalid",
			expErr: true,
		},
		{
			name:   "invalid hex address",
			sender: "0xinvalid",
			expErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgBurnCoins(tc.sender, sdk.NewCoins(sdk.NewCoin("ufuel", sdkmath.NewInt(100))))
			if tc.expErr {
				require.Panics(t, func() { msg.GetSigners() })
			} else {
				signers := msg.GetSigners()
				require.Len(t, signers, 1)
				require.Equal(t, tc.sender, signers[0].String())
			}
		})
	}
}
