package testsuite

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/cometbft/cometbft/crypto/merkle"
	"github.com/ethereum/go-ethereum/common"
	ethereumtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

// RunFuelStreamXProcess runs the FuelStreamX process for 1 proof generation.
func (s *E2ETestSuite) RunFuelStreamXProcess() (
	startBlock, targetBlock, headerHash, bridgeCommitment string, txReceipt *ethereumtypes.Receipt,
) {
	s.T().Log("starting fuelstreamx container...")
	var err error
	runOpts := dockertest.RunOptions{
		Name:         "fuelstreamx",
		Repository:   fuelStreamXDockerImageRepo,
		Tag:          fuelStreamXDockerImageTag,
		NetworkID:    s.dockerNetwork.Network.ID,
		PortBindings: map[docker.Port][]docker.PortBinding{},
		ExposedPorts: []string{},
		Env: []string{
			"RPC_URL=http://ethereum-node:8545",
			fmt.Sprintf("TENDERMINT_RPC_URL=http://%s:26657", s.Chain.validators[0].instanceName()),
			"CHAIN_ID=31337",
			fmt.Sprintf("CONTRACT_ADDRESS=%s", FUEL_STREAM_X_CONTRACT),
			fmt.Sprintf("PRIVATE_KEY=%s", s.GetEthPrivateKeyHex()),
			fmt.Sprintf("UPDATE_DELAY_BLOCKS=%d", UPDATE_DELAY_BLOCKS),
		},
	}

	s.fuelStreamXResource, err = s.dockerPool.RunWithOptions(
		&runOpts,
		noRestart,
	)
	s.Require().NoError(err)

	re1 := regexp.MustCompile(`\[[^]]*] updating header range starting (\d+), ending (\d+), header hash "([0-9a-fA-F]{64})", bridge commitment "([0-9a-fA-F]{64})"`)
	re2 := regexp.MustCompile(`\[[^]]*] Proof relayed successfully! Transaction Hash: (0x[a-fA-F0-9]{64})`)

	// Wait for the FuelStreamX process to respond
	s.Require().Eventually(
		func() bool {
			logs := s.logsByContainerID(s.fuelStreamXResource.Container.ID)

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			var ok1, ok2 bool
			for _, logStr := range strings.Split(logs, "\n") {

				matches := re1.FindStringSubmatch(logStr)
				// The first element represents the string captures, the rest are the matches obtained from the string
				if matches != nil && len(matches) == 5 {
					startBlock = matches[1]
					targetBlock = matches[2]
					headerHash = matches[3]
					bridgeCommitment = matches[4]
					ok1 = true
				}

				matches = re2.FindStringSubmatch(logStr)
				// The first element represents the string captures, the second is the tx hash
				if matches != nil && len(matches) > 1 {
					receipt, err := s.Chain.ethClient.TransactionReceipt(ctx, common.HexToHash(matches[1]))
					if err != nil {
						s.T().Logf("error retreiving transaction receipt %s", err)
						return false
					}
					s.Require().NotNil(receipt.Logs)

					txReceipt = receipt
					ok2 = true
				}

				if ok1 && ok2 {
					return true
				}
			}

			return false
		},
		1*time.Minute,
		1*time.Second,
		"FuelStreamX failed to respond",
	)

	s.T().Logf("FuelStreamX request successful!")

	// We only want 1 proof from the FuelStreamX process
	s.T().Logf("stopping FuelStreamX container...")
	s.Require().NoError(s.dockerPool.Purge(s.fuelStreamXResource))

	return
}

// AuntsToHashes takes aunts from a Merkle proof and converts them to 32-byte hashes.
func AuntsToHashes(proof merkle.Proof) (hashes []common.Hash) {
	hashes = make([]common.Hash, len(proof.Aunts))
	for i, aunt := range proof.Aunts {
		hashes[i] = common.BytesToHash(aunt)
	}
	return
}
