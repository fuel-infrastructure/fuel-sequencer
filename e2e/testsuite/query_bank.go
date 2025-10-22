package testsuite

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (s *E2ETestSuite) QueryBalance(
	ctx context.Context,
	address string,
	denom string,
) (*banktypes.QueryBalanceResponse, error) {
	queryClient := s.getGRPCClients().BankQueryClient

	res, err := queryClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: address,
		Denom:   denom,
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

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

func (s *E2ETestSuite) QuerySpendableBalance(
	ctx context.Context,
	address string,
	denom string,
) (*banktypes.QuerySpendableBalanceByDenomResponse, error) {
	queryClient := s.getGRPCClients().BankQueryClient

	res, err := queryClient.SpendableBalanceByDenom(ctx, &banktypes.QuerySpendableBalanceByDenomRequest{
		Address: address,
		Denom:   denom,
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *E2ETestSuite) QuerySpendableBalances(
	ctx context.Context,
	address string,
	pagination *query.PageRequest,
) (*banktypes.QuerySpendableBalancesResponse, error) {
	queryClient := s.getGRPCClients().BankQueryClient

	res, err := queryClient.SpendableBalances(ctx, &banktypes.QuerySpendableBalancesRequest{
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
	h, err := s.Chain.FuelSequencerHeight(ctx)
	s.Require().NoError(err)

	s.T().Log(fmt.Sprintf("Polling for balance %s of %s", balance.String(), address))

	doPoll := func(ctx context.Context, height uint64) error {
		bal, err := s.Chain.grpcClients.BankQueryClient.Balance(ctx, &banktypes.QueryBalanceRequest{
			Address: address,
			Denom:   balance.Denom,
		})
		if err != nil {
			return err
		}
		if !bal.Balance.Equal(balance) {
			return fmt.Errorf("balance (%s) does not match expected: (%s)", bal, balance.Amount.String())
		}
		return nil
	}

	bp := BlockPoller{CurrentHeight: s.Chain.FuelSequencerHeight, PollFunc: doPoll}
	err = bp.DoPoll(ctx, h, h+deltaBlocks)
	s.Require().NoError(err, "balance not found in expected number of blocks")
}

// PollForExactBalance polls until the full balance of the address matches the specified balances exactly.
// Unlike PollForMultipleBalances, this expects the full balance of the address to match the specified balance.
func (s *E2ETestSuite) PollForExactBalance(
	ctx context.Context, deltaBlocks uint64, address string, balances sdk.Coins,
) {
	h, err := s.Chain.FuelSequencerHeight(ctx)
	s.Require().NoError(err)

	s.T().Log(fmt.Sprintf("Polling for all balance %s of %s", balances.String(), address))

	doPoll := func(ctx context.Context, height uint64) error {
		bal, err := s.Chain.grpcClients.BankQueryClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
			Address: address,
		})
		if err != nil {
			return err
		}
		if !bal.Balances.Equal(balances) {
			return fmt.Errorf("balance (%s) does not match expected: (%s)", bal, balances.String())
		}
		return nil
	}

	bp := BlockPoller{CurrentHeight: s.Chain.FuelSequencerHeight, PollFunc: doPoll}
	err = bp.DoPoll(ctx, h, h+deltaBlocks)
	s.Require().NoError(err, "exact balance not found in expected number of blocks")
}

// PollForMultipleBalances polls until the balance matches exactly for all the individual balances specified.
// Unlike PollForExactBalance, this does not expect the full balance of the address to match the specified balance.
func (s *E2ETestSuite) PollForMultipleBalances(
	ctx context.Context, deltaBlocks uint64, address string, balances sdk.Coins,
) {
	h, err := s.Chain.FuelSequencerHeight(ctx)
	s.Require().NoError(err)

	s.T().Log(fmt.Sprintf("Polling for multiple balances %s of %s", balances.String(), address))

	doPoll := func(ctx context.Context, height uint64) error {
		bal, err := s.Chain.grpcClients.BankQueryClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
			Address: address,
		})
		if err != nil {
			return err
		}
		for _, expectedBalance := range balances {
			if !bal.Balances.AmountOf(expectedBalance.Denom).Equal(expectedBalance.Amount) {
				return fmt.Errorf("balance (%s) does not match expected: (%s)", bal, balances.String())
			}
		}
		return nil
	}

	bp := BlockPoller{CurrentHeight: s.Chain.FuelSequencerHeight, PollFunc: doPoll}
	err = bp.DoPoll(ctx, h, h+deltaBlocks)
	s.Require().NoError(err, "matching balances not found in expected number of blocks")
}
