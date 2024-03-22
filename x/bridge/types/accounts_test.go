package types_test

import (
	"encoding/hex"
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	_ "github.com/fuel-infrastructure/fuel-sequencer/app/apptesting" // ensure bech32 configs are set
	testutiltypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"github.com/stretchr/testify/require"
)

func TestGenerateSequencerAccountFromEthereumAddress(t *testing.T) {

	accAddress, err := types.GenerateSequencerAddressFromEthereumAddress(testutiltypes.TestEthAddr1Str)
	require.NoError(t, err)

	require.Equal(t, testutiltypes.TestSeqAddr1Str, accAddress.String())
}

func TestGenerateSequencerAccountFromEthereumAddressFromBz(t *testing.T) {

	ethAddress := testutiltypes.TestEthAddr1Str
	ethAddressBz, err := hex.DecodeString(ethAddress[2:])
	require.NoError(t, err)

	accAddress, err := types.GenerateSequencerAddressFromEthereumAddressFromBz(ethAddressBz)
	require.NoError(t, err)

	require.Equal(t, testutiltypes.TestSeqAddr1Str, accAddress.String())
}

func TestEthOwnedBaseAccountSetSequenceErrors(t *testing.T) {
	baseAcc := authtypes.NewBaseAccountWithAddress(testutiltypes.TestSeqAddr1)
	acc := types.NewEthOwnedBaseAccount(baseAcc, testutiltypes.TestEthAddr1Str)
	require.ErrorContains(t, acc.SetSequence(1), "cannot set sequence number for eth owned account")
	require.ErrorContains(t, acc.SetSequence(2), "cannot set sequence number for eth owned account")
}

func TestEthOwnedBaseAccountSetPubkeyErrors(t *testing.T) {
	_, pk, _ := testdata.KeyTestPubAddr()

	acc := types.NewEthOwnedBaseAccount(&authtypes.BaseAccount{}, "")
	require.ErrorContains(t, acc.SetPubKey(pk), "cannot set public key for eth owned account")
	require.ErrorContains(t, acc.SetPubKey(pk), "cannot set public key for eth owned account")
}

func TestEthOwnedContinuousVestingAccountSetSequenceErrors(t *testing.T) {
	acc := types.NewEthOwnedContinuousVestingAccount(&vestingtypes.ContinuousVestingAccount{}, "")
	require.ErrorContains(t, acc.SetSequence(1), "cannot set sequence number for eth owned continuous vesting account")
	require.ErrorContains(t, acc.SetSequence(2), "cannot set sequence number for eth owned continuous vesting account")
}

func TestEthOwnedContinuousVestingAccountSetPubkeyErrors(t *testing.T) {
	_, pk, _ := testdata.KeyTestPubAddr()

	acc := types.NewEthOwnedContinuousVestingAccount(&vestingtypes.ContinuousVestingAccount{}, "")
	require.ErrorContains(t, acc.SetPubKey(pk), "cannot set public key for eth owned continuous vesting account")
	require.ErrorContains(t, acc.SetPubKey(pk), "cannot set public key for eth owned continuous vesting account")
}
