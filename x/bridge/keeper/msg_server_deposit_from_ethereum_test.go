package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
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
		name           string
		msgs           []*types.MsgDepositFromEthereum
		toAcc          *sdk.AccAddress
		fromAcc        *sdk.AccAddress
		expFromBalance sdkmath.Int
		expToBalance   sdkmath.Int
		expSupplyDelta *types.SupplyDeltaInfo
		expGovBal      sdkmath.Int
		isToEthOwned   bool
		isFromEthOwned bool
	}{
		{
			name: "successful - deposit - mint to Recipient address",
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
			name: "successful - deposit - mint to Recipient address twice",
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
			name: "successful - deposit - mint to address Recipient == Depositor",
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
			name: "successful - deposit - mint to address Recipient == Sequencer(Depositor)",
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
			name: "successful - deposit - mint to Depositor address",
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
			name: "successful - deposit - mint to Depositor address twice",
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
			name: "failure - deposit - lockup failed to parse - mint to governance",
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
			name: "failure - deposit - from address failed to parse - mint to governance",
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
			name: "failure - deposit - bad vesting duration - mint to governance",
			msgs: []*types.MsgDepositFromEthereum{
				testtypes.TestEvent6Msg,
			},
			fromAcc:        &fromAccOne,
			toAcc:          nil,
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
			name: "failure - deposit - bad to bech32 address - mint to governance",
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
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Get the message server
			msgServer := keeper.NewMsgServerImpl(s.App.BridgeKeeper)

			for _, msg := range tc.msgs {
				_, err = msgServer.DepositFromEthereum(s.Ctx(), msg)
				s.Require().NoError(err)
			}

			// Confirm that the balances were changed as specified. This indicates that the AuthorizedEvents were
			// executed successfully
			if tc.fromAcc != nil {
				actualFromBalance := s.App.BankKeeper.GetBalance(s.Ctx(), *tc.fromAcc, "ufuel")
				s.Require().Equal(tc.expFromBalance, actualFromBalance.Amount)

				// Verify if the account is ETH owned
				account := s.App.AccountKeeper.GetAccount(s.Ctx(), *tc.fromAcc)
				_, ok := account.(types.EthOwnedAccountI)
				s.Require().Equal(tc.isFromEthOwned, ok)
			}
			if tc.toAcc != nil {
				actualToBalance := s.App.BankKeeper.GetBalance(s.Ctx(), *tc.toAcc, "ufuel")
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
			actualGovBal := s.App.BankKeeper.GetBalance(s.Ctx(), govAddr, "ufuel")
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

	// Trigger the processing of the Ethereum events
	s.Require().Panics(func() {
		_, _ = msgServer.DepositFromEthereum(s.Ctx(), msg)
	})

	// Verify the governance address balance is as expected
	actualGovBal := s.App.BankKeeper.GetBalance(s.Ctx(), govAddr, "ufuel")
	s.Require().Equal(sdkmath.ZeroInt(), actualGovBal.Amount)
}
