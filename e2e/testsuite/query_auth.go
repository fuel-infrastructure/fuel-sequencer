package testsuite

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// QueryModuleAccountAddress returns the sdk.AccAddress of a given module name.
func (s *E2ETestSuite) QueryModuleAccountAddress(ctx context.Context, moduleName string) (sdk.AccAddress, error) {
	authClient := s.getGRPCClients().AuthQueryClient

	resp, err := authClient.ModuleAccountByName(ctx, &authtypes.QueryModuleAccountByNameRequest{
		Name: moduleName,
	})
	if err != nil {
		return nil, err
	}

	cfg := encodingConfig

	var account authtypes.AccountI
	if err := cfg.InterfaceRegistry.UnpackAny(resp.Account, &account); err != nil {
		return nil, err
	}
	moduleAccount, ok := account.(authtypes.ModuleAccountI)
	if !ok {
		return nil, fmt.Errorf("failed to cast account: %T as ModuleAccount", moduleAccount)
	}

	return moduleAccount.GetAddress(), nil
}
