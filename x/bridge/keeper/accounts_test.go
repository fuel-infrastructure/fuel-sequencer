package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

const (
	firstAccNumber   = 7 // this is not zero due to module accounts
	firstAccSequence = 0
)

func (s *KeeperTestSuite) TestGetSequencerAccountForEthereumAddress() {

	ethAddr1Str := "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"
	//ethAddr2Str := "0x0Ac72d9E87B39DAAa81e4F3F29Ce8c45B2bE5fA9"

	// corresponds to ethAddress1
	seqAddr1Str := "fuelsequencer13tch2uhman7dhjjphmx9uwx7kvg2kqfj5y56hsmljlv93pgma5vqyks99k"
	seqAddr1 := sdk.MustAccAddressFromBech32(seqAddr1Str)

	token100 := sdk.NewCoins(sdk.NewInt64Coin("token", 100))

	type accountValidator func(acc sdk.AccountI) bool

	matchesBaseAcc := func(baseAcc *authtypes.BaseAccount) accountValidator {
		return func(acc sdk.AccountI) bool {
			bAcc, ok := acc.(*authtypes.BaseAccount)
			return ok &&
				(bAcc.GetAddress().Equals(baseAcc.GetAddress())) &&
				((bAcc.PubKey == nil && baseAcc.PubKey == nil) || (bAcc.GetPubKey().Equals(baseAcc.GetPubKey()))) &&
				(bAcc.GetAccountNumber() == baseAcc.GetAccountNumber()) &&
				(bAcc.GetSequence() == baseAcc.GetSequence())
		}
	}

	//matchesContinuousVestingAccount := func(vestingAcc *vestingtypes.ContinuousVestingAccount) accountValidator {
	//	return func(acc sdk.AccountI) bool {
	//		vAcc, ok := acc.(*vestingtypes.ContinuousVestingAccount)
	//		return ok &&
	//			(vAcc.StartTime == vestingAcc.StartTime) &&
	//			(vAcc.OriginalVesting.Equal(vestingAcc.OriginalVesting)) &&
	//			(vAcc.DelegatedFree.Equal(vestingAcc.DelegatedFree)) &&
	//			(vAcc.DelegatedVesting.Equal(vestingAcc.DelegatedVesting)) &&
	//			(vAcc.EndTime != vestingAcc.EndTime) &&
	//			matchesBaseAcc(vestingAcc.BaseAccount)(vAcc.BaseAccount)
	//	}
	//}

	type args struct {
		ethAddress      string
		vestingDuration time.Duration
		totalCoins      sdk.Coins
	}
	testsCases := []struct {
		name                string
		precreateAccount    sdk.AccountI
		args                args
		isAccountAsExpected func(sdk.AccountI) bool
		expectErrMsg        string
	}{
		{
			name: "new account with no vesting and no coins => valid base account",
			args: args{
				ethAddress:      ethAddr1Str,
				vestingDuration: 0,
				totalCoins:      nil,
			},
			isAccountAsExpected: matchesBaseAcc(
				authtypes.NewBaseAccount(seqAddr1, nil, firstAccNumber, firstAccSequence),
			),
		},
		{
			name: "new account with no vesting and some coins => valid base account",
			args: args{
				ethAddress:      ethAddr1Str,
				vestingDuration: 0,
				totalCoins:      token100,
			},
			isAccountAsExpected: matchesBaseAcc(
				authtypes.NewBaseAccount(seqAddr1, nil, firstAccNumber, firstAccSequence),
			),
		},
	}

	for _, tc := range testsCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			accAddress, err := s.App.BridgeKeeper.GetSequencerAccountForEthereumAddress(
				s.Ctx(), tc.args.ethAddress, tc.args.vestingDuration, tc.args.totalCoins,
			)

			if tc.expectErrMsg != "" {
				s.Require().ErrorContains(err, tc.expectErrMsg)
				return
			}
			s.Require().NoError(err)

			account := s.App.AccountKeeper.GetAccount(s.Ctx(), accAddress)
			s.Require().True(tc.isAccountAsExpected(account))
		})
	}
}
