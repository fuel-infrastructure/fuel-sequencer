package testsuite

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (s *E2ETestSuite) QueryAllBalances(
	ctx context.Context,
	address string,
	pagination *query.PageRequest,
) (*banktypes.QueryAllBalancesResponse, error) {
	queryClient := s.getGRPCClients().BankQueryClient

	res, err := queryClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
		Address:    address,
		Pagination: pagination,
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *E2ETestSuite) QueryBankParams(ctx context.Context) *banktypes.Params {
	queryClient := s.getGRPCClients().BankQueryClient
	res, err := queryClient.Params(ctx, &banktypes.QueryParamsRequest{})
	s.Require().NoError(err)

	return &res.Params
}

func (s *E2ETestSuite) QueryBankSendEnabled(ctx context.Context) []*banktypes.SendEnabled {
	queryClient := s.getGRPCClients().BankQueryClient
	res, err := queryClient.SendEnabled(ctx, &banktypes.QuerySendEnabledRequest{})
	s.Require().NoError(err)

	return res.SendEnabled
}

// PollForBalance polls until the balance matches
func (s *E2ETestSuite) PollForBalance(
	ctx context.Context, deltaBlocks uint64, address string, balance sdk.Coin,
) {
	h, err := s.chain.FuelSequencerHeight(ctx)
	s.Require().NoError(err)

	s.T().Log(fmt.Sprintf("Polling for balance %s of %s", balance.String(), address))

	doPoll := func(ctx context.Context, height uint64) (any, error) {
		bal, err := s.chain.grpcClients.BankQueryClient.Balance(ctx, &banktypes.QueryBalanceRequest{
			Address: address,
			Denom:   balance.Denom,
		})
		if err != nil {
			return nil, err
		}
		if !bal.Balance.Equal(balance) {
			return nil, fmt.Errorf("balance (%d) does not match expected: (%d)", bal, balance.Amount.Int64())
		}
		return nil, nil
	}

	bp := BlockPoller[any]{CurrentHeight: s.chain.FuelSequencerHeight, PollFunc: doPoll}
	_, err = bp.DoPoll(ctx, h, h+deltaBlocks)
	s.Require().NoError(err, "balance not found in expected number of blocks")
}

// PollForExactBalance polls until the full balance of the address matches the specified balances exactly.
// Unlike PollForMultipleBalances, this expects the full balance of the address to match the specified balance.
func (s *E2ETestSuite) PollForExactBalance(
	ctx context.Context, deltaBlocks uint64, address string, balances sdk.Coins,
) {
	h, err := s.chain.FuelSequencerHeight(ctx)
	s.Require().NoError(err)

	s.T().Log(fmt.Sprintf("Polling for all balance %s of %s", balances.String(), address))

	doPoll := func(ctx context.Context, height uint64) (any, error) {
		bal, err := s.chain.grpcClients.BankQueryClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
			Address: address,
		})
		if err != nil {
			return nil, err
		}
		if !bal.Balances.Equal(balances) {
			return nil, fmt.Errorf("balance (%s) does not match expected: (%s)", bal, balances.String())
		}
		return nil, nil
	}

	bp := BlockPoller[any]{CurrentHeight: s.chain.FuelSequencerHeight, PollFunc: doPoll}
	_, err = bp.DoPoll(ctx, h, h+deltaBlocks)
	s.Require().NoError(err, "exact balance not found in expected number of blocks")
}

// PollForMultipleBalances polls until the balance matches exactly for all the individual balances specified.
// Unlike PollForExactBalance, this does not expect the full balance of the address to match the specified balance.
func (s *E2ETestSuite) PollForMultipleBalances(
	ctx context.Context, deltaBlocks uint64, address string, balances sdk.Coins,
) {
	h, err := s.chain.FuelSequencerHeight(ctx)
	s.Require().NoError(err)

	s.T().Log(fmt.Sprintf("Polling for multiple balances %s of %s", balances.String(), address))

	doPoll := func(ctx context.Context, height uint64) (any, error) {
		bal, err := s.chain.grpcClients.BankQueryClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
			Address: address,
		})
		if err != nil {
			return nil, err
		}
		for _, expectedBalance := range balances {
			if !bal.Balances.AmountOf(expectedBalance.Denom).Equal(expectedBalance.Amount) {
				return nil, fmt.Errorf("balance (%s) does not match expected: (%s)", bal, balances.String())
			}
		}
		return nil, nil
	}

	bp := BlockPoller[any]{CurrentHeight: s.chain.FuelSequencerHeight, PollFunc: doPoll}
	_, err = bp.DoPoll(ctx, h, h+deltaBlocks)
	s.Require().NoError(err, "matching balances not found in expected number of blocks")
}
