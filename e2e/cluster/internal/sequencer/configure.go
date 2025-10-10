// Package cluster provides functionality for setting up and managing a distributed
// network of Fuel Sequencer validator nodes. It handles binary building, configuration,
// deployment and management of the network.
package sequencer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/fuel-infrastructure/fuel-sequencer/app"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	sequencingtypes "github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"

	"cosmossdk.io/math"
	cmconfig "github.com/cometbft/cometbft/config"
	cmjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cosmos/cosmos-sdk/server"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

// sequencer represents a validator node setup in the Fuel Sequencer network
type sequencer struct {
	chain *testsuite.Chain

	// keys are the FuelSequencer wallets derived from the mnemonics, with hex versions of the addresses.
	// This is filled-in later on in SetupTest, once the address codec has been initialised.
	keys []*testsuite.SequencerKey
}

// ConfigureNetwork sets up the initial network configuration including chain creation,
// genesis initialization, and validator configurations.
// Returns an error if any configuration step fails.
func ConfigureNetwork(logger *zap.SugaredLogger,
	dataDir string, locally func(logger *zap.SugaredLogger, cmd string, args ...string) error,
	peerIPs []string,
) error {
	l := logger.Named("Configure")

	// Initialize chain with defined number of nodes
	chain, err := testsuite.NewNamedChain(ChainName, dataDir, len(Mnemonics))
	if err != nil {
		return fmt.Errorf("failed to create chain: %w", err)
	}

	if _, err := os.Stat(chain.ConfigDir()); !os.IsNotExist(err) {
		l.Warnw("data directory already exists - deleting existing configuration...", "network", ChainName, "path", chain.ConfigDir())
		locally(l, "rm", "-rf", chain.ConfigDir())
	}

	l.Infow("setting up data for new chain...", "name", ChainName, "path", chain.ConfigDir())

	s := &sequencer{chain: chain}

	// Derive and output the Sequencer keys with the hex and bech32 representation of the addresses.
	for i, mnemonic := range Mnemonics {
		key := testsuite.MustNewSequencerKeyFromMnemonic(mnemonic)
		l.Infow("generated sequencer key", "index", i, "mnemonic", mnemonic, "acc", key.AddressSeq, "val", key.ValAddressSeq, "hex", key.AddressHex)
		s.keys = append(s.keys, key)
	}

	err = s.initNodes()
	if err != nil {
		return fmt.Errorf("failed to initialise nodes: %w", err)
	}

	err = s.initGenesis()
	if err != nil {
		return fmt.Errorf("failed to initialise genesis: %w", err)
	}

	err = s.initValidatorConfigs(peerIPs)
	if err != nil {
		return fmt.Errorf("failed to initialise validator configs: %w", err)
	}

	// err = s.checkValidators()
	// if err != nil {
	// 	return fmt.Errorf("failed to run validators: %w", err)
	// }

	return nil
}

// initNodes initializes validator nodes and their genesis accounts.
// Returns an error if node initialization fails.
func (s *sequencer) initNodes() error {
	err := s.chain.CreateAndInitFuelSequencerValidators(Mnemonics)
	if err != nil {
		return fmt.Errorf("failed to setup nodes from genesis: %w", err)
	}

	// initialize a genesis file for the first validator
	val0ConfigDir := s.chain.Validators[0].ConfigDir()
	for _, val := range s.chain.Validators {
		if err := testsuite.AddGenesisAccount(val0ConfigDir, "", initBalanceCoin.String(), val.Address()); err != nil {
			return fmt.Errorf("failed to add genesis account: %w", err)
		}
	}

	// copy the genesis file to the remaining validators
	for _, val := range s.chain.Validators[1:] {
		err := testsuite.CopyFile(
			filepath.Join(val0ConfigDir, "config", "genesis.json"),
			filepath.Join(val.ConfigDir(), "config", "genesis.json"),
		)
		if err != nil {
			return fmt.Errorf("failed to copy genesis file: %w", err)
		}
	}

	return nil
}

// initGenesis initializes and customizes the genesis state for all validators.
// Configures governance, mint, bank, bridge and sequencing parameters.
// Returns an error if genesis initialization fails.
func (s *sequencer) initGenesis() error {
	cdc := testsuite.TestCdc

	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(s.chain.Validators[0].ConfigDir())
	config.Moniker = s.chain.Validators[0].Moniker

	genFilePath := config.GenesisFile()
	appGenState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFilePath)
	if err != nil {
		return fmt.Errorf("failed to get genesis state from genesis file: %w", err)
	}

	// set short voting periods to allow gov proposals in tests
	var govGenState govtypesv1.GenesisState
	err = cdc.UnmarshalJSON(appGenState[govtypes.ModuleName], &govGenState)
	if err != nil {
		return fmt.Errorf("failed to unmarshal gov genesis state: %w", err)
	}

	votingPeriod := governanceVotingPeriod
	govGenState.Params.VotingPeriod = &votingPeriod
	govGenState.Params.ExpeditedVotingPeriod = &votingPeriod
	govGenState.Params.MinDeposit = sdk.Coins{{Denom: bridgeDenom, Amount: math.OneInt()}}
	govGenState.Params.ExpeditedMinDeposit = sdk.Coins{{Denom: bridgeDenom, Amount: math.OneInt()}}
	bz, err := cdc.MarshalJSON(&govGenState)
	if err != nil {
		return fmt.Errorf("failed to marshal gov genesis state: %w", err)
	}
	appGenState[govtypes.ModuleName] = bz

	// set mint denom
	var mintGenState minttypes.GenesisState
	err = cdc.UnmarshalJSON(appGenState[minttypes.ModuleName], &mintGenState)
	if err != nil {
		return fmt.Errorf("failed to unmarshal mint genesis state: %w", err)
	}
	mintGenState.Params.InflationMax = math.LegacyZeroDec()
	mintGenState.Params.InflationMin = math.LegacyZeroDec()
	mintGenState.Params.InflationRateChange = math.LegacyZeroDec()
	mintGenState.Minter.Inflation = math.LegacyZeroDec()
	bz, err = cdc.MarshalJSON(&mintGenState)
	if err != nil {
		return fmt.Errorf("failed to marshal mint genesis state: %w", err)
	}
	appGenState[minttypes.ModuleName] = bz

	// TODO: genesis supply will be incorrect if we add more accounts
	var bankGenState banktypes.GenesisState
	err = cdc.UnmarshalJSON(appGenState[banktypes.ModuleName], &bankGenState)
	if err != nil {
		return fmt.Errorf("failed to unmarshal bank genesis state: %w", err)
	}
	genesisSupply := initBalanceCoin.Amount.MulRaw(int64(len(s.chain.Validators)))
	bankGenState.Supply = sdk.NewCoins(sdk.NewCoin(bridgeDenom, genesisSupply))
	bz, err = cdc.MarshalJSON(&bankGenState)
	if err != nil {
		return fmt.Errorf("failed to marshal bank genesis state: %w", err)
	}
	appGenState[banktypes.ModuleName] = bz

	// Set the vesting start time to the genesis time to ensure that vesting tests are not dependent on specific times,
	// making them consistently pass regardless of when they are executed.
	vestingStartingTime := genDoc.GenesisTime

	var bridgeGenState bridgetypes.GenesisState
	err = cdc.UnmarshalJSON(appGenState[bridgetypes.ModuleName], &bridgeGenState)
	if err != nil {
		return fmt.Errorf("failed to unmarshal bridge genesis state: %w", err)
	}
	bridgeGenState.Params.BridgeDenom = bridgeDenom
	bridgeGenState.Params.SupplyDeltaPeriod = supplyDeltaPeriod
	bridgeGenState.Params.VestingStartTime = vestingStartingTime
	bridgeGenState.Params.BridgeDenomTotalSupply = bridgeDenomTotalSupply
	bz, err = cdc.MarshalJSON(&bridgeGenState)
	if err != nil {
		return fmt.Errorf("failed to marshal bridge genesis state: %w", err)
	}
	appGenState[bridgetypes.ModuleName] = bz

	var sequencingGenState sequencingtypes.GenesisState
	err = cdc.UnmarshalJSON(appGenState[sequencingtypes.ModuleName], &sequencingGenState)
	if err != nil {
		return fmt.Errorf("failed to unmarshal sequencing genesis state: %w", err)
	}
	sequencingGenState.Params.MaxBlobSizeBytes = blobMaxBytes
	sequencingGenState.Params.SequencerTxMaxBytes = blockMaxGas
	loadTestSequencerConfigs(&sequencingGenState)
	bz, err = cdc.MarshalJSON(&sequencingGenState)
	if err != nil {
		return fmt.Errorf("failed to marshal sequencing genesis state: %w", err)
	}
	appGenState[sequencingtypes.ModuleName] = bz

	var genUtilGenState genutiltypes.GenesisState
	err = cdc.UnmarshalJSON(appGenState[genutiltypes.ModuleName], &genUtilGenState)
	if err != nil {
		return fmt.Errorf("failed to unmarshal genutil genesis state: %w", err)
	}

	// generate genesis txs
	genTxs := make([]json.RawMessage, len(s.chain.Validators))
	for i, val := range s.chain.Validators {
		createValmsg, err := val.BuildCreateValidatorMsg(val.InstanceName(), initStakedCoin)
		if err != nil {
			return fmt.Errorf("failed to build create validator message: %w", err)
		}

		signedTx, err := val.SignMsg(createValmsg)
		if err != nil {
			return fmt.Errorf("failed to sign create validator message: %w", err)
		}

		txRaw, err := cdc.MarshalJSON(signedTx)
		if err != nil {
			return fmt.Errorf("failed to marshal create validator message: %w", err)
		}

		genTxs[i] = txRaw
	}

	genUtilGenState.GenTxs = genTxs

	bz, err = cdc.MarshalJSON(&genUtilGenState)
	if err != nil {
		return fmt.Errorf("failed to marshal genutil genesis state: %w", err)
	}
	appGenState[genutiltypes.ModuleName] = bz

	// // Apply any genesis overrides
	// if s.GenesisOverrides != nil {
	// 	err = (*s.GenesisOverrides)(cdc, appGenState)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to apply genesis overrides: %w", err)
	// 	}
	// }

	// serialize genesis state
	bz, err = json.MarshalIndent(appGenState, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal genesis state: %w", err)
	}

	genDoc.AppState = bz

	bz, err = cmjson.MarshalIndent(genDoc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal genesis document: %w", err)
	}

	// write the updated genesis file to each validator
	for _, val := range s.chain.Validators {
		err = testsuite.WriteFile(filepath.Join(val.ConfigDir(), "config", "genesis.json"), bz)
		if err != nil {
			return fmt.Errorf("failed to write genesis file: %w", err)
		}
	}

	return nil
}

// initValidatorConfigs initializes validator-specific configurations including
// P2P settings, RPC endpoints, and application parameters.
// Returns an error if validator configuration fails.
func (s *sequencer) initValidatorConfigs(peerIPs []string) error {
	for i, val := range s.chain.Validators {
		cmCfgPath := filepath.Join(val.ConfigDir(), "config", "config.toml")

		vpr := viper.New()
		vpr.SetConfigFile(cmCfgPath)
		if err := vpr.ReadInConfig(); err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}

		valConfig := &cmconfig.Config{}
		if err := vpr.Unmarshal(valConfig); err != nil {
			return fmt.Errorf("failed to unmarshal config file: %w", err)
		}

		valConfig.Mempool.MaxTxBytes = mempoolMaxTxBytes
		valConfig.Mempool.MaxTxsBytes = mempoolMaxTxsBytes
		valConfig.P2P.ListenAddress = "tcp://0.0.0.0:26656"
		valConfig.P2P.AddrBookStrict = false
		valConfig.P2P.ExternalAddress = fmt.Sprintf("%s:%d", peerIPs[i], 26656)
		valConfig.RPC.ListenAddress = "tcp://0.0.0.0:26657"
		valConfig.RPC.MaxBodyBytes = int64(blockMaxGas)
		valConfig.StateSync.Enable = false
		valConfig.LogLevel = "info"
		valConfig.Instrumentation.Prometheus = true

		// speed up blocks
		valConfig.Consensus.TimeoutCommit = 5 * time.Second
		valConfig.Consensus.TimeoutPropose = 3 * time.Second

		// overwrite using CometBFT configs
		loadTestCometBFTConfigs(valConfig)

		var peers []string

		for j := 0; j < len(s.chain.Validators); j++ {
			if i == j {
				continue
			}

			peer := s.chain.Validators[j]
			peerID := fmt.Sprintf("%s@%s:26656", peer.NodeKey.ID(), peerIPs[j])
			peers = append(peers, peerID)
		}

		valConfig.P2P.PersistentPeers = strings.Join(peers, ",")
		valConfig.Moniker = val.InstanceName()
		cmconfig.WriteConfigFile(cmCfgPath, valConfig)

		// set application configuration
		appCfgPath := filepath.Join(val.ConfigDir(), "config", "app.toml")

		customAppTemplate, customAppConfig := app.DefaultCustomAppConfig()
		appConfig := customAppConfig.(app.CustomAppConfig)
		appConfig.Mempool.MaxTxs = 0 // unlimited txs
		appConfig.QueryGasLimit = 0  // unlimited gas
		appConfig.API.Enable = true
		appConfig.API.Address = "tcp://0.0.0.0:1317"
		appConfig.API.RPCMaxBodyBytes = uint(blockMaxGas)
		appConfig.GRPC.Address = "0.0.0.0:9090"
		appConfig.Pruning = "nothing"
		appConfig.MinGasPrices = fmt.Sprintf("%s%s", minGasPrices, bridgeDenom)
		appConfig.CommitmentsConfig.ApiEnabled = true
		appConfig.CommitmentsConfig.MaxQueryRange = 4096
		appConfig.Telemetry.Enabled = true
		appConfig.Telemetry.PrometheusRetentionTime = 60 // 1 minute
		appConfig.SidecarConfig.Enabled = false

		srvconfig.SetConfigTemplate(customAppTemplate)
		srvconfig.WriteConfigFile(appCfgPath, appConfig)
	}

	return nil
}
