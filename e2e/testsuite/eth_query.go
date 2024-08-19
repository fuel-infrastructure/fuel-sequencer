package testsuite

import (
	"context"
	"math/big"
	"strings"

	errorsmod "cosmossdk.io/errors"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

func (s *E2ETestSuite) QueryEthereumEthBalance(
	ctx context.Context, account common.Address, blockNumber *big.Int,
) (*big.Int, error) {
	return s.Chain.ethClient.BalanceAt(ctx, account, blockNumber)
}

func (s *E2ETestSuite) QueryEthereumEthPendingBalance(
	ctx context.Context, account common.Address,
) (*big.Int, error) {
	return s.Chain.ethClient.PendingBalanceAt(ctx, account)
}

func (s *E2ETestSuite) QueryEthereumErc20Balance(
	ctx context.Context, account, contract common.Address, contractABI string,
) (*big.Int, error) {

	query := ethereum.CallMsg{
		To:   &contract,
		Data: PackBalanceOfERC20Token(contractABI, account),
	}

	result, err := s.Chain.ethClient.CallContract(ctx, query, nil)
	if err != nil {
		return nil, err
	}

	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		panic(errorsmod.Wrap(err, "bad ABI definition in code"))
	}

	balance := new(big.Int)
	err = parsedABI.UnpackIntoInterface(&balance, BalanceOfQueryName, result)
	if err != nil {
		return nil, err
	}

	return balance, nil
}

// QueryEthereumMigratedTokenBalance queries the migrated token (FUEL V1) balance of the specified account.
func (s *E2ETestSuite) QueryEthereumMigratedTokenBalance(
	ctx context.Context, account common.Address,
) (*big.Int, error) {
	return s.QueryEthereumErc20Balance(ctx, account, MigratedTokenContractAddress, MigratedTokenContractABI)
}

// QueryEthereumTokenBalance queries the token (FUEL V2) balance of the specified account.
func (s *E2ETestSuite) QueryEthereumTokenBalance(
	ctx context.Context, account common.Address,
) (*big.Int, error) {
	return s.QueryEthereumErc20Balance(ctx, account, TokenContractAddress, TokenContractABI)
}

// QueryEthereumAddressHasRole queries whether an address has the specified role.
func (s *E2ETestSuite) QueryEthereumAddressHasRole(
	ctx context.Context, account, contract common.Address, contractABI string, role common.Hash,
) (bool, error) {

	query := ethereum.CallMsg{
		To:   &contract,
		Data: PackHasRole(contractABI, role, account),
	}

	result, err := s.Chain.ethClient.CallContract(ctx, query, nil)
	if err != nil {
		return false, err
	}

	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		panic(errorsmod.Wrap(err, "bad ABI definition in code"))
	}

	var hasRole bool
	err = parsedABI.UnpackIntoInterface(&hasRole, HasRoleQueryName, result)
	if err != nil {
		return false, err
	}

	return hasRole, nil
}

// QueryEthereumAddressHasRole_FuelStreamXContract calls QueryEthereumAddressHasRole for the FuelStreamX contract.
func (s *E2ETestSuite) QueryEthereumAddressHasRole_FuelStreamXContract(
	ctx context.Context, account common.Address, role common.Hash,
) (bool, error) {
	return s.QueryEthereumAddressHasRole(ctx, account, FuelStreamXContractAddress, FuelStreamXContractABI, role)
}
