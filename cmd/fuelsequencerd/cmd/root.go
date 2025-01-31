package cmd

import (
	"fmt"
	"os"
	"strings"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/client/v2/autocli"
	clientv2keyring "cosmossdk.io/client/v2/autocli/keyring"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/registry"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/flags"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtxconfig "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/fuel-infrastructure/fuel-sequencer/app"

	authv1 "cosmossdk.io/api/cosmos/auth/module/v1"
	stakingv1 "cosmossdk.io/api/cosmos/staking/module/v1"
	basedepinject "cosmossdk.io/x/accounts/defaults/base/depinject"
	lockupdepinject "cosmossdk.io/x/accounts/defaults/lockup/depinject"
	multisigdepinject "cosmossdk.io/x/accounts/defaults/multisig/depinject"
)

// NewRootCmd creates a new root command for fuelsequencerd. It is called once in the main function.
func NewRootCmd() *cobra.Command {
	app.InitSDKConfig()

	var (
		txConfigOpts  tx.ConfigOptions
		autoCliOpts   autocli.AppOptions
		moduleManager *module.Manager
		clientCtx     client.Context
	)

	if err := depinject.Inject(
		depinject.Configs(
			app.AppConfig(),
			depinject.Supply(
				log.NewNopLogger(),
				servertypes.AppOptions(AppOptionsMap{
					flags.FlagHome: app.DefaultNodeHome, // otherwise x/upgrade creates a data/ folder at the CWD
				}),
			),
			depinject.Provide(
				ProvideClientContext,
				ProvideKeyring,

				multisigdepinject.ProvideAccount,
				basedepinject.ProvideAccount,
				lockupdepinject.ProvideAllLockupAccounts,

				// provide base account options
				basedepinject.ProvideSecp256K1PubKey,
			),
		),
		&txConfigOpts,
		&autoCliOpts,
		&moduleManager,
		&clientCtx,
	); err != nil {
		panic(err)
	}

	rootCmd := &cobra.Command{
		Use:           app.Name + "d",
		Short:         fmt.Sprintf("Start %s node", app.PrettyName),
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// set the default command outputs
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

			clientCtx = clientCtx.WithCmdContext(cmd.Context())
			clientCtx, err := client.ReadPersistentCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			// We can override default client config template and configs.
			customClientTemplate := ""
			customConfig := interface{}(nil)

			clientCtx, err = config.CreateClientConfig(clientCtx, customClientTemplate, customConfig)
			if err != nil {
				return err
			}

			// This needs to go after ReadFromClientConfig, as that function
			// sets the RPC client needed for SIGN_MODE_TEXTUAL.
			txConfigOpts.EnabledSignModes = append(txConfigOpts.EnabledSignModes, signingv1beta1.SignMode_SIGN_MODE_TEXTUAL)
			txConfigOpts.TextualCoinMetadataQueryFn = authtxconfig.NewGRPCCoinMetadataQueryFn(clientCtx)
			txConfigWithTextual, err := tx.NewTxConfigWithOptions(
				codec.NewProtoCodec(clientCtx.InterfaceRegistry),
				txConfigOpts,
			)
			if err != nil {
				return err
			}

			clientCtx = clientCtx.WithTxConfig(txConfigWithTextual)
			if err := client.SetCmdClientContextHandler(clientCtx, cmd); err != nil {
				return err
			}

			customAppTemplate, customAppConfig := app.DefaultCustomAppConfig()
			customCMTConfig := app.InitCometBFTConfig()

			return server.InterceptConfigsPreRunHandler(cmd, customAppTemplate, customAppConfig, customCMTConfig)
		},
	}

	initRootCmd(rootCmd, moduleManager)

	nodeCmds := nodeservice.NewNodeCommands()
	autoCliOpts.ModuleOptions = make(map[string]*autocliv1.ModuleOptions)
	autoCliOpts.ModuleOptions[nodeCmds.Name()] = nodeCmds.AutoCLIOptions()

	overwriteFlagDefaults(rootCmd, map[string]string{
		flags.FlagChainID:        strings.ReplaceAll(app.Name, "-", ""),
		flags.FlagKeyringBackend: "test",
	})

	if err := autoCliOpts.EnhanceRootCommand(rootCmd); err != nil {
		panic(err)
	}

	return rootCmd
}

func overwriteFlagDefaults(c *cobra.Command, defaults map[string]string) {
	set := func(s *pflag.FlagSet, key, val string) {
		if f := s.Lookup(key); f != nil {
			f.DefValue = val

			//nolint:errcheck
			f.Value.Set(val)
		}
	}
	for key, val := range defaults {
		set(c.Flags(), key, val)
		set(c.PersistentFlags(), key, val)
	}
	for _, c := range c.Commands() {
		overwriteFlagDefaults(c, defaults)
	}
}

func ProvideClientContext(
	appCodec codec.Codec,
	interfaceRegistry codectypes.InterfaceRegistry,
	txConfigOpts tx.ConfigOptions,
	legacyAmino registry.AminoRegistrar,
	addressCodec address.Codec,
	validatorAddressCodec address.ValidatorAddressCodec,
	consensusAddressCodec address.ConsensusAddressCodec,
	authConfig *authv1.Module,
	stakingConfig *stakingv1.Module,
) client.Context {
	var err error

	amino, ok := legacyAmino.(*codec.LegacyAmino)
	if !ok {
		panic("ProvideClientContext requires a *codec.LegacyAmino instance")
	}

	clientCtx := client.Context{}.
		WithCodec(appCodec).
		WithInterfaceRegistry(interfaceRegistry).
		WithLegacyAmino(amino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithAddressCodec(addressCodec).
		WithValidatorAddressCodec(validatorAddressCodec).
		WithConsensusAddressCodec(consensusAddressCodec).
		WithHomeDir(app.DefaultNodeHome).
		WithViper(app.Name). // env variable prefix
		WithAddressPrefix(authConfig.Bech32Prefix).
		WithValidatorPrefix(stakingConfig.Bech32PrefixValidator)

	// We can override default client config template and configs.
	customClientTemplate := ""
	customConfig := interface{}(nil)

	clientCtx, err = config.CreateClientConfig(clientCtx, customClientTemplate, customConfig)
	if err != nil {
		panic(err)
	}

	// textual is enabled by default, we need to re-create the tx config grpc instead of bank keeper.
	txConfigOpts.TextualCoinMetadataQueryFn = authtxconfig.NewGRPCCoinMetadataQueryFn(clientCtx)
	txConfig, err := tx.NewTxConfigWithOptions(clientCtx.Codec, txConfigOpts)
	if err != nil {
		panic(err)
	}
	clientCtx = clientCtx.WithTxConfig(txConfig)

	return clientCtx
}

func ProvideKeyring(clientCtx client.Context, addressCodec address.Codec) (clientv2keyring.Keyring, error) {
	kb, err := client.NewKeyringFromBackend(clientCtx, clientCtx.Keyring.Backend())
	if err != nil {
		return nil, err
	}

	return keyring.NewAutoCLIKeyring(kb, addressCodec)
}
