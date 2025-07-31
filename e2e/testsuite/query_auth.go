package testsuite

import (
	"context"
	"fmt"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// QueryModuleAccountAddress returns the sdk.AccAddress of a given module name.
func (s *E2ETestSuite) QueryModuleAccountAddress(ctx context.Context, moduleName string) (sdktypes.AccAddress, error) {
	authClient := s.getGRPCClients().AuthQueryClient

	resp, err := authClient.ModuleAccountByName(ctx, &authtypes.QueryModuleAccountByNameRequest{
		Name: moduleName,
	})
	if err != nil {
		return nil, err
	}

	cfg := encodingConfig

	var account sdktypes.AccountI
	if err := cfg.InterfaceRegistry.UnpackAny(resp.Account, &account); err != nil {
		return nil, err
	}
	moduleAccount, ok := account.(sdktypes.ModuleAccountI)
	if !ok {
		return nil, fmt.Errorf("failed to cast account: %T as ModuleAccount", moduleAccount)
	}

	return moduleAccount.GetAddress(), nil
}

// QueryEthOwnedContinuousVestingAccount returns an EthOwnedContinuousVestingAccount corresponding to the given address
// if no errors.
func (s *E2ETestSuite) QueryEthOwnedContinuousVestingAccount(
	ctx context.Context, address string,
) (*bridgetypes.EthOwnedContinuousVestingAccount, error) {
	authClient := s.getGRPCClients().AuthQueryClient
	resp, err := authClient.Account(ctx, &authtypes.QueryAccountRequest{Address: address})
	if err != nil {
		return nil, err
	}

	cfg := encodingConfig

	var account sdktypes.AccountI
	if err := cfg.InterfaceRegistry.UnpackAny(resp.Account, &account); err != nil {
		return nil, err
	}
	ethOwnedContinuousVestingAccount, ok := account.(*bridgetypes.EthOwnedContinuousVestingAccount)
	if !ok {
		return nil, fmt.Errorf(
			"failed to cast account: %T as EthOwnedContinuousVestingAccount", ethOwnedContinuousVestingAccount,
		)
	}

	return ethOwnedContinuousVestingAccount, nil
}

// QueryEthOwnedMultiContinuousVestingAccount returns an EthOwnedMultiContinuousVestingAccount corresponding to the
// given address if no errors.
func (s *E2ETestSuite) QueryEthOwnedMultiContinuousVestingAccount(
	ctx context.Context, address string,
) (*bridgetypes.EthOwnedMultiContinuousVestingAccount, error) {
	authClient := s.getGRPCClients().AuthQueryClient
	resp, err := authClient.Account(ctx, &authtypes.QueryAccountRequest{Address: address})
	if err != nil {
		return nil, err
	}

	cfg := encodingConfig

	var account sdktypes.AccountI
	if err := cfg.InterfaceRegistry.UnpackAny(resp.Account, &account); err != nil {
		return nil, err
	}
	ethOwnedMultiContinuousVestingAccount, ok := account.(*bridgetypes.EthOwnedMultiContinuousVestingAccount)
	if !ok {
		return nil, fmt.Errorf(
			"failed to cast account: %T as EthOwnedMultiContinuousVestingAccount", ethOwnedMultiContinuousVestingAccount,
		)
	}

	return ethOwnedMultiContinuousVestingAccount, nil
}

// QueryEthOwnedBaseAccount returns an EthOwnedBaseAccount corresponding to the given address if no errors.
func (s *E2ETestSuite) QueryEthOwnedBaseAccount(
	ctx context.Context, address string,
) (*bridgetypes.EthOwnedBaseAccount, error) {
	authClient := s.getGRPCClients().AuthQueryClient
	resp, err := authClient.Account(ctx, &authtypes.QueryAccountRequest{Address: address})
	if err != nil {
		return nil, err
	}

	cfg := encodingConfig

	var account sdktypes.AccountI
	if err := cfg.InterfaceRegistry.UnpackAny(resp.Account, &account); err != nil {
		return nil, err
	}
	ethOwnedBaseAccount, ok := account.(*bridgetypes.EthOwnedBaseAccount)
	if !ok {
		return nil, fmt.Errorf(
			"failed to cast account: %T as EthOwnedBaseAccount", ethOwnedBaseAccount,
		)
	}

	return ethOwnedBaseAccount, nil
}

// QueryBaseAccount returns a BaseAccount corresponding to the given address if no errors.
func (s *E2ETestSuite) QueryBaseAccount(ctx context.Context, address string) (*authtypes.BaseAccount, error) {
	authClient := s.getGRPCClients().AuthQueryClient
	resp, err := authClient.Account(ctx, &authtypes.QueryAccountRequest{Address: address})
	if err != nil {
		return nil, err
	}

	cfg := encodingConfig

	var account sdktypes.AccountI
	if err := cfg.InterfaceRegistry.UnpackAny(resp.Account, &account); err != nil {
		return nil, err
	}
	baseAccount, ok := account.(*authtypes.BaseAccount)
	if !ok {
		return nil, fmt.Errorf("failed to cast account: %T as BaseAccount", baseAccount)
	}

	return baseAccount, nil
}

// QueryContinuousVestingAccount returns a ContinuousVestingAccount corresponding to the given address if no errors.
func (s *E2ETestSuite) QueryContinuousVestingAccount(
	ctx context.Context, address string,
) (*vestingtypes.ContinuousVestingAccount, error) {
	authClient := s.getGRPCClients().AuthQueryClient
	resp, err := authClient.Account(ctx, &authtypes.QueryAccountRequest{Address: address})
	if err != nil {
		return nil, err
	}

	cfg := encodingConfig

	var account sdktypes.AccountI
	if err := cfg.InterfaceRegistry.UnpackAny(resp.Account, &account); err != nil {
		return nil, err
	}
	continuousVestingAccount, ok := account.(*vestingtypes.ContinuousVestingAccount)
	if !ok {
		return nil, fmt.Errorf("failed to cast account: %T as ContinuousVestingAccount", continuousVestingAccount)
	}

	return continuousVestingAccount, nil
}
