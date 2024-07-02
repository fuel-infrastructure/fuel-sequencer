package keeper_test

import (
	"strings"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *KeeperTestSuite) TestWithdrawToEthereum() {
	withdrawer := testtypes.TestFrom1
	withdrawerSeqLower := strings.ToLower(testtypes.TestFrom1Seq)
	withdrawerLower := strings.ToLower(testtypes.TestFrom1)
	withdrawerSeqUpper := strings.ToUpper(testtypes.TestFrom1Seq)
	withdrawerUpper := strings.ToUpper(testtypes.TestFrom1)

	receiver := testtypes.TestTo1
	receiverLower := strings.ToLower(testtypes.TestTo1)
	receiverUpper := strings.ToUpper(testtypes.TestTo1)

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
			"successfully withdraw to Ethereum - hex from - lowercase addresses",
			types.MsgWithdrawToEthereum{
				From:   withdrawerLower,
				To:     receiverLower,
				Amount: sdk.NewCoin(bridgeDenom, amount),
			},
			&types.MsgWithdrawToEthereumResponse{
				Nonce:  math.NewInt(1),
				From:   withdrawerLower,
				To:     receiverLower,
				Amount: sdk.NewCoin(bridgeDenom, amount),
			},
			&types.SupplyDeltaInfo{
				LastSupply: amount,
				Offset:     amount,
				ToReport:   amount,
			},
			true,
			"",
		},
		{
			"successfully withdraw to Ethereum - hex from - uppercase addresses",
			types.MsgWithdrawToEthereum{
				From:   withdrawerUpper,
				To:     receiverUpper,
				Amount: sdk.NewCoin(bridgeDenom, amount),
			},
			&types.MsgWithdrawToEthereumResponse{
				Nonce:  math.NewInt(1),
				From:   withdrawerLower, // changed to lowercase
				To:     receiverLower,   // changed to lowercase
				Amount: sdk.NewCoin(bridgeDenom, amount),
			},
			&types.SupplyDeltaInfo{
				LastSupply: amount,
				Offset:     amount,
				ToReport:   amount,
			},
			true,
			"",
		},
		{
			"successfully withdraw to Ethereum - bech32 from - lowercase addresses",
			types.MsgWithdrawToEthereum{
				From:   withdrawerSeqLower,
				To:     receiverLower,
				Amount: sdk.NewCoin(bridgeDenom, amount),
			},
			&types.MsgWithdrawToEthereumResponse{
				Nonce:  math.NewInt(1),
				From:   withdrawerSeqLower,
				To:     receiverLower,
				Amount: sdk.NewCoin(bridgeDenom, amount),
			},
			&types.SupplyDeltaInfo{
				LastSupply: amount,
				Offset:     amount,
				ToReport:   amount,
			},
			true,
			"",
		},
		{
			"successfully withdraw to Ethereum - bech32 from - uppercase addresses",
			types.MsgWithdrawToEthereum{
				From:   withdrawerSeqUpper,
				To:     receiverUpper,
				Amount: sdk.NewCoin(bridgeDenom, amount),
			},
			&types.MsgWithdrawToEthereumResponse{
				Nonce:  math.NewInt(1),
				From:   withdrawerSeqLower, // changed to lowercase
				To:     receiverLower,      // changed to lowercase
				Amount: sdk.NewCoin(bridgeDenom, amount),
			},
			&types.SupplyDeltaInfo{
				LastSupply: amount,
				Offset:     amount,
				ToReport:   amount,
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
				Offset:     math.ZeroInt(),
				ToReport:   math.ZeroInt(),
			},
			true,
			"invalid token denom",
		},
		{
			"error decoding from address",
			types.MsgWithdrawToEthereum{
				From:   "invalid-address",
				To:     receiver,
				Amount: sdk.NewCoin(bridgeDenom, amount),
			},
			nil,
			&types.SupplyDeltaInfo{
				LastSupply: math.ZeroInt(),
				Offset:     math.ZeroInt(),
				ToReport:   math.ZeroInt(),
			},
			false,
			"failed to decode from address",
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
				Offset:     math.ZeroInt(),
				ToReport:   math.ZeroInt(),
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
			_ = s.App.BridgeKeeper.SetParams(s.Ctx(), params)

			// Get the SupplyDeltaInfo
			if tc.fundAccounts {
				fromAcc, err := s.App.BridgeKeeper.GetAddressCodec().StringToBytes(tc.msg.From)
				s.Require().NoError(err)

				// Fund withdrawer account with the bridge denom
				s.FundAcc(s.Ctx(), fromAcc, sdk.NewCoins(tc.msg.Amount))

				// Update the supplyDeltaInfo since we've minted
				s.App.BridgeKeeper.UpdateSupplyDeltaInfoWithNewDelta(s.Ctx(), s.App.BankKeeper)
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
