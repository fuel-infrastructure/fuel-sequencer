package keeper_test

import (
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/ethereum/go-ethereum/common"
	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *KeeperTestSuite) TestDepositFromEthereum() {
	govAddr := s.App.AccountKeeper.GetModuleAddress(govtypes.ModuleName)

	// These accounts correspond to the from and to addresses of the DepositEvent
	fromAccOne, _ := s.App.BridgeKeeper.GenerateSequencerAddressFromEthereumAddress(testtypes.TestFrom3)
	toAccOne, err := s.App.BridgeKeeper.GenerateSequencerAddressFromEthereumAddress(testtypes.TestTo3)
	s.Require().NoError(err)

	fromAccTwo, _ := s.App.BridgeKeeper.GenerateSequencerAddressFromEthereumAddress(testtypes.TestFrom2)

	testCases := []struct {
		name             string
		blockedAddresses []string
		msgs             []*types.MsgDepositFromEthereum
		toAcc            *sdk.AccAddress
		fromAcc          *sdk.AccAddress
		expFromBalance   sdkmath.Int
		expToBalance     sdkmath.Int
		expSupplyDelta   *types.SupplyDeltaInfo
		expGovBal        sdkmath.Int
		isToEthOwned     bool
		isFromEthOwned   bool
		expErrMsg        string
	}{
		{
			name: "successful - mint to Recipient address",
			msgs: []*types.MsgDepositFromEthereum{
				testtypes.TestEvent1Msg,
			},
			fromAcc:        &fromAccOne,
			toAcc:          &toAccOne,
			expFromBalance: sdkmath.NewInt(0),
			expToBalance:   sdkmath.NewInt(102),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-102),
			},
			expGovBal:      sdkmath.ZeroInt(),
			isToEthOwned:   false,
			isFromEthOwned: false,
		},
		{
			name: "successful - mint to Recipient address twice",
			msgs: []*types.MsgDepositFromEthereum{
				testtypes.TestEvent1Msg,
				testtypes.TestEvent1Msg,
			},
			fromAcc:        &fromAccOne,
			toAcc:          &toAccOne,
			expFromBalance: sdkmath.NewInt(0),
			expToBalance:   sdkmath.NewInt(204),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-204),
			},
			expGovBal:      sdkmath.ZeroInt(),
			isToEthOwned:   false,
			isFromEthOwned: false,
		},
		{
			name: "successful - mint to address Recipient == Depositor",
			msgs: []*types.MsgDepositFromEthereum{
				testtypes.TestEvent10Msg,
			},
			fromAcc:        &fromAccOne,
			toAcc:          &fromAccOne,
			expFromBalance: sdkmath.NewInt(100), // Same balance Depositor == Recipient
			expToBalance:   sdkmath.NewInt(100),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-100),
			},
			expGovBal:      sdkmath.ZeroInt(),
			isToEthOwned:   true,
			isFromEthOwned: true,
		},
		{
			name: "successful - mint to address Recipient == Sequencer(Depositor)",
			msgs: []*types.MsgDepositFromEthereum{
				testtypes.TestEvent11Msg,
			},
			fromAcc:        &fromAccOne,
			toAcc:          &fromAccOne,
			expFromBalance: sdkmath.NewInt(100), // Same balance Depositor == Recipient
			expToBalance:   sdkmath.NewInt(100),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-100),
			},
			expGovBal:      sdkmath.ZeroInt(),
			isToEthOwned:   true,
			isFromEthOwned: true,
		},
		{
			name: "successful - mint to Depositor address",
			msgs: []*types.MsgDepositFromEthereum{
				testtypes.TestEvent3Msg,
			},
			fromAcc:        &fromAccTwo,
			toAcc:          nil,
			expFromBalance: sdkmath.NewInt(101),
			expToBalance:   sdkmath.NewInt(0),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-101),
			},
			expGovBal:      sdkmath.ZeroInt(),
			isToEthOwned:   false,
			isFromEthOwned: true,
		},
		{
			name: "successful - mint to Depositor address twice",
			msgs: []*types.MsgDepositFromEthereum{
				testtypes.TestEvent3Msg,
				testtypes.TestEvent3Msg,
			},
			fromAcc:        &fromAccTwo,
			toAcc:          nil,
			expFromBalance: sdkmath.NewInt(202),
			expToBalance:   sdkmath.NewInt(0),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-202),
			},
			expGovBal:      sdkmath.ZeroInt(),
			isToEthOwned:   false,
			isFromEthOwned: true,
		},
		{
			name: "failure - lockup failed to parse - mint to governance",
			msgs: []*types.MsgDepositFromEthereum{
				testtypes.TestEvent4Msg,
			},
			fromAcc:        &fromAccOne,
			toAcc:          &toAccOne,
			expFromBalance: sdkmath.NewInt(0),
			expToBalance:   sdkmath.NewInt(0),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-102),
			},
			expGovBal:      sdkmath.NewInt(102),
			isToEthOwned:   false,
			isFromEthOwned: false,
		},
		{
			name: "failure - depositor address failed to parse - mint to governance",
			msgs: []*types.MsgDepositFromEthereum{
				testtypes.TestEvent5Msg,
			},
			fromAcc:        &fromAccOne,
			toAcc:          &toAccOne,
			expFromBalance: sdkmath.NewInt(0),
			expToBalance:   sdkmath.NewInt(0),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-102),
			},
			expGovBal:      sdkmath.NewInt(102),
			isToEthOwned:   false,
			isFromEthOwned: false,
		},
		{
			name: "failure - bad recipient bech32 address - mint to governance",
			msgs: []*types.MsgDepositFromEthereum{
				testtypes.TestEvent7Msg,
			},
			fromAcc:        &fromAccOne,
			toAcc:          &toAccOne,
			expFromBalance: sdkmath.NewInt(0),
			expToBalance:   sdkmath.NewInt(0),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-102),
			},
			expGovBal:      sdkmath.NewInt(102),
			isToEthOwned:   false,
			isFromEthOwned: false,
		},
		{
			name:             "failure - depositor address (in bech32) is blocked - mint to governance",
			blockedAddresses: []string{sdk.AccAddress(common.FromHex(testtypes.TestEvent1Msg.Depositor)).String()},
			msgs: []*types.MsgDepositFromEthereum{
				testtypes.TestEvent1Msg,
			},
			fromAcc:        &fromAccOne,
			toAcc:          &toAccOne,
			expFromBalance: sdkmath.NewInt(0),
			expToBalance:   sdkmath.NewInt(0),
			expSupplyDelta: &types.SupplyDeltaInfo{
				Offset: sdkmath.NewInt(-102),
			},
			expGovBal:      sdkmath.NewInt(102),
			isToEthOwned:   false,
			isFromEthOwned: false,
		},
		{
			name: "failure - invalid authority address",
			msgs: []*types.MsgDepositFromEthereum{
				{Authority: "fuelsequencer17w0adeg64ky0daxwd2ugyuneellmjgnx5dpmtz"},
			},
			expErrMsg: "invalid authority",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Block addresses, maybe
			if len(tc.blockedAddresses) > 0 {
				params := s.App.BridgeKeeper.GetParams(s.Ctx())
				params.AdditionalBlockedAddresses = append(params.AdditionalBlockedAddresses, tc.blockedAddresses...)
				s.Require().NoError(s.App.BridgeKeeper.SetParams(s.Ctx(), params))
			}

			// Get the message server
			msgServer := keeper.NewMsgServerImpl(s.App.BridgeKeeper)

			// Set the VestingStartTime since by default it's not valid
			params := s.App.BridgeKeeper.GetParams(s.Ctx())
			params.VestingStartTime = time.Now()
			s.Require().NoError(s.App.BridgeKeeper.SetParams(s.Ctx(), params))

			for _, msg := range tc.msgs {
				_, err = msgServer.DepositFromEthereum(s.Ctx(), msg)
				if tc.expErrMsg != "" {
					s.Require().ErrorContains(err, tc.expErrMsg)
					return
				}
				s.Require().NoError(err)
			}

			// Confirm that the balances were changed as specified. This indicates that the AuthorizedEvents were
			// executed successfully
			if tc.fromAcc != nil {
				actualFromBalance := s.App.BankKeeper.GetBalance(s.Ctx(), *tc.fromAcc, types.DefaultBridgeDenom)
				s.Require().Equal(tc.expFromBalance, actualFromBalance.Amount)

				// Verify if the account is ETH owned
				account := s.App.AccountKeeper.GetAccount(s.Ctx(), *tc.fromAcc)
				_, ok := account.(types.EthOwnedAccountI)
				s.Require().Equal(tc.isFromEthOwned, ok)
			}
			if tc.toAcc != nil {
				actualToBalance := s.App.BankKeeper.GetBalance(s.Ctx(), *tc.toAcc, types.DefaultBridgeDenom)
				s.Require().Equal(tc.expToBalance, actualToBalance.Amount)

				// Verify if the account is ETH owned
				account := s.App.AccountKeeper.GetAccount(s.Ctx(), *tc.toAcc)
				_, ok := account.(types.EthOwnedAccountI)
				s.Require().Equal(tc.isToEthOwned, ok)
			}

			// Verify the supply delta info was updated
			if tc.expSupplyDelta != nil {
				actualSupplyDelta := s.App.BridgeKeeper.MustGetSupplyDeltaInfo(s.Ctx())
				s.Require().Equal(tc.expSupplyDelta.Offset, actualSupplyDelta.Offset)
			}

			// Verify the governance address balance is as expected
			actualGovBal := s.App.BankKeeper.GetBalance(s.Ctx(), govAddr, types.DefaultBridgeDenom)
			s.Require().Equal(tc.expGovBal, actualGovBal.Amount)
		})
	}
}

func (s *KeeperTestSuite) TestDepositFromEthereum_AmountParseFailure() {
	govAddr := s.App.AccountKeeper.GetModuleAddress(govtypes.ModuleName)

	msg := testtypes.TestEvent8Msg
	s.SetupTest() // Reset the test suite

	// Get the message server
	msgServer := keeper.NewMsgServerImpl(s.App.BridgeKeeper)

	// Get balances before
	allBalancesBefore := s.App.BankKeeper.GetAccountsBalances(s.Ctx())

	// Trigger the processing of the Ethereum events
	_, _ = msgServer.DepositFromEthereum(s.Ctx(), msg)

	// Verify that all balances are unchanged
	allBalancesAfter := s.App.BankKeeper.GetAccountsBalances(s.Ctx())
	s.Require().EqualValues(allBalancesBefore, allBalancesAfter)

	// Verify the governance address balance is still zero
	actualGovBal := s.App.BankKeeper.GetBalance(s.Ctx(), govAddr, types.DefaultBridgeDenom)
	s.Require().True(actualGovBal.Amount.IsZero())
}
