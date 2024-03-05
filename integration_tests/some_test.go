package integration_tests

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *IntegrationTestSuite) TestAuction() {
	s.Run("Bring up chain", func() {
		val := s.chain.validators[0]
		kb, err := val.keyring()
		s.Require().NoError(err)
		val0ClientCtx, err := s.chain.clientContext("tcp://localhost:26657", &kb, "val", val.address())
		s.Require().NoError(err)
		//auctionQueryClient := types.NewQueryClient(val0ClientCtx)
		_ = val0ClientCtx
	})
}

func balanceOfDenom(balances sdk.Coins, denom string) (found bool, balance sdk.Coin) {
	for _, balance := range balances {
		if balance.Denom == denom {
			return true, balance
		}
	}

	return false, sdk.NewCoin(denom, math.ZeroInt())
}
