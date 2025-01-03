package testsuite

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	cmconfig "github.com/cometbft/cometbft/config"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/fuel-infrastructure/fuel-sequencer/app"
	"github.com/spf13/viper"
)

func (s *E2ETestSuite) initFuelSequencerValidatorConfigs() {
	for i, val := range s.Chain.Validators {
		cmCfgPath := filepath.Join(val.ConfigDir(), "config", "config.toml")

		vpr := viper.New()
		vpr.SetConfigFile(cmCfgPath)
		s.Require().NoError(vpr.ReadInConfig())

		valConfig := &cmconfig.Config{}
		s.Require().NoError(vpr.Unmarshal(valConfig))

		valConfig.P2P.ListenAddress = "tcp://0.0.0.0:26656"
		valConfig.P2P.AddrBookStrict = false
		valConfig.P2P.ExternalAddress = fmt.Sprintf("%s:%d", val.InstanceName(), 26656)
		valConfig.RPC.ListenAddress = "tcp://0.0.0.0:26657"
		valConfig.StateSync.Enable = false
		valConfig.LogLevel = "info"
		valConfig.Instrumentation.Prometheus = true

		// speed up blocks
		valConfig.Consensus.TimeoutCommit = 1 * time.Second
		valConfig.Consensus.TimeoutPropose = 1 * time.Second

		var peers []string

		for j := 0; j < len(s.Chain.Validators); j++ {
			if i == j {
				continue
			}

			peer := s.Chain.Validators[j]
			peerID := fmt.Sprintf("%s@%s%d:26656", peer.NodeKey.ID(), peer.Moniker, j)
			peers = append(peers, peerID)
		}

		valConfig.P2P.PersistentPeers = strings.Join(peers, ",")

		cmconfig.WriteConfigFile(cmCfgPath, valConfig)

		// set application configuration
		appCfgPath := filepath.Join(val.ConfigDir(), "config", "app.toml")

		customAppTemplate, customAppConfig := app.DefaultCustomAppConfig()
		appConfig := customAppConfig.(app.CustomAppConfig)
		appConfig.API.Enable = true
		appConfig.API.Address = "tcp://0.0.0.0:1317"
		appConfig.GRPC.Address = "0.0.0.0:9090"
		appConfig.Pruning = "nothing"
		appConfig.MinGasPrices = fmt.Sprintf("%s%s", minGasPrices, BridgeDenom)
		appConfig.CommitmentsConfig.ApiEnabled = true
		appConfig.CommitmentsConfig.MaxQueryRange = 4096
		appConfig.Telemetry.Enabled = true
		appConfig.Telemetry.PrometheusRetentionTime = 60 // 1 minute

		srvconfig.SetConfigTemplate(customAppTemplate)
		srvconfig.WriteConfigFile(appCfgPath, appConfig)
	}
}
