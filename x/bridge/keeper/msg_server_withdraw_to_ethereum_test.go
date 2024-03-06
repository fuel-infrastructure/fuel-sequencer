package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *KeeperTestSuite) TestWithdrawToEthereum() {
	withdrawer := s.TestAccs[0].String()
	receiver := "0x95222290DD7278Aa3Ddd389Cc1E1d165CC4BAfe6"
	bridgeDenom := "fuel"
	invalidDenom := "invalidDenom"
	amount := math.NewInt(200)

	testCases := []struct {
		name               string
		msg                types.MsgWithdrawToEthereum
		msgResponse        *types.MsgWithdrawToEthereumResponse
		expSupplyDeltaInfo *types.SupplyDeltaInfo
		fundAccounts       bool
		expErrMsg          string
	}{
		{
			"successfully withdraw to Ethereum",
			types.MsgWithdrawToEthereum{
				From:   withdrawer,
				To:     receiver,
				Amount: sdk.NewCoin(bridgeDenom, amount),
			},
			&types.MsgWithdrawToEthereumResponse{
				Nonce:  math.NewInt(1),
				From:   withdrawer,
				To:     receiver,
				Amount: sdk.NewCoin(bridgeDenom, amount),
			},
			&types.SupplyDeltaInfo{
				LastSupply: math.ZeroInt(),
				Delta:      math.ZeroInt(),
				Offset:     math.ZeroInt(),
			},
			true,
			"",
		},
		{
			"error with invalid denom",
			types.MsgWithdrawToEthereum{
				From:   withdrawer,
				To:     receiver,
				Amount: sdk.NewCoin(invalidDenom, amount),
			},
			nil,
			&types.SupplyDeltaInfo{
				LastSupply: math.ZeroInt(),
				Delta:      math.ZeroInt(),
				Offset:     math.ZeroInt(),
			},
			true,
			"invalid token denom",
		},
		{
			"error burning coins from address",
			types.MsgWithdrawToEthereum{
				From:   withdrawer,
				To:     receiver,
				Amount: sdk.NewCoin(bridgeDenom, amount),
			},
			nil,
			&types.SupplyDeltaInfo{
				LastSupply: math.ZeroInt(),
				Delta:      math.ZeroInt(),
				Offset:     math.ZeroInt(),
			},
			false,
			"failed to burn bridge tokens",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Get the message server
			msgServer := keeper.NewMsgServerImpl(s.App.BridgeKeeper)

			// Set bridge parameters
			params := types.DefaultParams()
			params.BridgeDenom = bridgeDenom
			s.App.BridgeKeeper.SetParams(s.Ctx(), params)

			// Get the SupplyDeltaInfo
			if tc.fundAccounts {
				// Fund withdrawer account with the bridge denom
				s.FundAcc(s.Ctx(), sdk.MustAccAddressFromBech32(tc.msg.From), sdk.NewCoins(tc.msg.Amount))
				// Update the supplyDeltaInfo since we've minted
				s.App.BridgeKeeper.UpdatedSupplyDeltaInfoWithNewDelta(s.Ctx(), s.App.BankKeeper)
			}

			response, err := msgServer.WithdrawToEthereum(s.Ctx(), &tc.msg)
			s.Require().Equal(tc.msgResponse, response)
			if len(tc.expErrMsg) > 0 {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			}

			afterSupplyDeltaInfo := s.App.BridgeKeeper.MustGetSupplyDeltaInfo(s.Ctx())
			s.Require().Equal(tc.expSupplyDeltaInfo, &afterSupplyDeltaInfo)
		})
	}
}
