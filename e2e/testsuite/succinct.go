package testsuite

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	cmbytes "github.com/cometbft/cometbft/libs/bytes"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

type SuccinctXProof struct {
	ChainId    uint64 `json:"chain_id"`
	To         string `json:"to"`
	Calldata   string `json:"calldata"`
	FunctionId string `json:"function_id"`
	Input      string `json:"input"`
	Proof      string `json:"proof"`
	Output     string `json:"output"`
}

// Note: The operator is only active for 1 proof generation
func (s *E2ETestSuite) RunSuccinctXOperatorMockApi() (string, string, string) {
	s.T().Log("starting SuccinctX operator container...")
	var err error
	runOpts := dockertest.RunOptions{
		Name:         "succinctX-operator",
		Repository:   succinctXOperatorDockerImageRepo,
		Tag:          succinctXOperatorDockerImageTag,
		NetworkID:    s.dockerNetwork.Network.ID,
		PortBindings: map[docker.Port][]docker.PortBinding{},
		ExposedPorts: []string{},
		Env: []string{
			"ETHEREUM_RPC_URL=http://ethereum:8545",
			fmt.Sprintf("TENDERMINT_RPC_URL=http://%s:26657", s.chain.validators[0].instanceName()),
			"SUCCINCT_RPC_URL=http://localhost:1234", // Can be anything
			"SUCCINCT_API_KEY=",                      // Can be anything
			"MOCK_SUCCINCT_SERVER=true",              // Mocking Succinct API server
			"CHAIN_ID=31337",
			fmt.Sprintf("CONTRACT_ADDRESS=%s", FUEL_STREAM_X_CONTRACT),
			fmt.Sprintf("NEXT_HEADER_FUNCTION_ID=%s", NEXT_HEADER_FUNCTION_ID),
			fmt.Sprintf("HEADER_RANGE_FUNCTION_ID=%s", HEADER_RANGE_FUNCTION_ID),
			"POST_DELAY_MINUTES=0", // No delays
			"LOCAL_PROVE_MODE=false",
			"LOCAL_RELAY_MODE=false",
		},
	}

	s.succinctOperatorResource, err = s.dockerPool.RunWithOptions(
		&runOpts,
		noRestart,
	)
	s.Require().NoError(err)

	var requestId string
	var startBlock string
	var targetBlock string

	// Wait for the Operator node to response
	s.Require().Eventually(
		func() bool {
			logs := s.logsByContainerID(s.succinctOperatorResource.Container.ID)

			for _, logStr := range strings.Split(logs, "\n") {
				re := regexp.MustCompile(`\[[^]]*] Header range request submitted: (\d+), start block: (\d+), target block: (\d+)`)

				matches := re.FindStringSubmatch(logStr)
				// The first element represents the string captures, the rest are the digits obtained from the string
				if matches != nil && len(matches) == 4 {
					requestId = matches[1]
					startBlock = matches[2]
					targetBlock = matches[3]
					return true
				}
			}

			return false
		},
		1*time.Minute,
		10*time.Second,
		"SuccinctX operator failed to respond",
	)

	s.T().Logf("SuccinctX operator request successful!")

	// We only want 1 proof from the operator
	s.T().Logf("stopping SuccinctX operator container...")
	s.Require().NoError(s.dockerPool.Purge(s.succinctOperatorResource))

	return requestId, startBlock, targetBlock
}

// Note: The relayer only submits 1 proof to Ethereum
func (s *E2ETestSuite) RunSuccinctXRelayerMockApi(
	requestId string,
	startBlock uint64,
	targetBlock uint64,
	latestHeaderHash cmbytes.HexBytes,
) {
	// Re-create the proof output
	commitment, err := s.chain.BridgeCommitment(s.Ctx(), startBlock, targetBlock)
	s.Require().NoError(err)

	startBlockBytes, err := To8PaddedHexBytes(startBlock)
	s.Require().NoError(err)

	targetBlockBytes, err := To8PaddedHexBytes(targetBlock)
	s.Require().NoError(err)

	padded32TargetBlockBytes, err := To32PaddedHexBytes(targetBlock)
	s.Require().NoError(err)

	proof := SuccinctXProof{
		ChainId: 31337,
		To:      FUEL_STREAM_X_CONTRACT,
		// To create the signature, you can use "cast calldata "commitHeaderRange(uint64)" 6"
		Calldata:   "0x89daae09" + cmbytes.HexBytes(padded32TargetBlockBytes).String(),
		FunctionId: HEADER_RANGE_FUNCTION_ID,
		Input:      "0x" + cmbytes.HexBytes(startBlockBytes).String() + latestHeaderHash.String() + cmbytes.HexBytes(targetBlockBytes).String(),
		Proof:      "0xbaaaaa", // Can be anything, not used
		Output:     "0x" + cmbytes.HexBytes(padded32TargetBlockBytes).String() + commitment.String(),
	}
	proofBz, err := json.Marshal(proof)
	s.Require().NoError(err)

	// Random dir
	dirPath, err := os.MkdirTemp(os.TempDir(), "fuelsequencer-e2e-testnet")
	s.Require().NoError(err)
	s.Require().NoError(writeFile(filepath.Join(dirPath, "output_1.json"), proofBz))

	s.T().Log("starting SuccinctX relayer container...")
	runOpts := dockertest.RunOptions{
		Name:         "succinctX-relayer",
		Repository:   succinctXRelayerDockerImageRepo,
		Tag:          succinctXRelayerDockerImageTag,
		NetworkID:    s.dockerNetwork.Network.ID,
		PortBindings: map[docker.Port][]docker.PortBinding{},
		ExposedPorts: []string{},
		Cmd:          []string{"--", "--request-id", requestId},
		Mounts: []string{
			fmt.Sprintf("%s/:%s", dirPath, "/app/proofs"),
		},
		Env: []string{
			"ETHEREUM_RPC_URL=http://ethereum:8545",
			"PRIVATE_KEY=0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80",
			"SUCCINCT_RPC_URL=http://localhost:1234", // Can be anything
			"SUCCINCT_API_KEY=",                      // Can be anything
			fmt.Sprintf("GATEWAY_ADDRESS=%s", GATEWAY_CONTRACT),
			"LOCAL_PROVE_MODE=true",
			"LOCAL_RELAY_MODE=true",
		},
	}

	s.succinctRelayerResource, err = s.dockerPool.RunWithOptions(
		&runOpts,
		noRestart,
	)
	s.Require().NoError(err)

	// Wait for the Relayer node to response
	match := "Relayed successfully!"
	s.Require().Eventually(
		func() bool {
			logs := s.logsByContainerID(s.succinctOperatorResource.Container.ID)

			return strings.Contains(logs, match)
		},
		1*time.Minute,
		10*time.Second,
		"SuccinctX relayer failed to respond",
	)

	// We only want 1 proof submitted from the relayer
	s.T().Logf("stopping SuccinctX operator container...")
	s.Require().NoError(s.dockerPool.Purge(s.succinctOperatorResource))
}

// -------------- TEMP

// padBytes Pad bytes to given length
func padBytes(byt []byte, length int) ([]byte, error) {
	l := len(byt)
	if l > length {
		return nil, fmt.Errorf(
			"cannot pad bytes because length of bytes array: %d is greater than given length: %d",
			l,
			length,
		)
	}
	if l == length {
		return byt, nil
	}
	tmp := make([]byte, length)
	copy(tmp[length-l:], byt)
	return tmp, nil
}

// To32PaddedHexBytes takes a number and returns its hex representation padded to 32 bytes.
// Used to mimic the result of `abi.encode(number)` in Ethereum.
func To32PaddedHexBytes(number uint64) ([]byte, error) {
	hexRepresentation := strconv.FormatUint(number, 16)
	// Make sure hex representation has even length.
	// The `strconv.FormatUint` can return odd length hex encodings.
	// For example, `strconv.FormatUint(10, 16)` returns `a`.
	// Thus, we need to pad it.
	if len(hexRepresentation)%2 == 1 {
		hexRepresentation = "0" + hexRepresentation
	}
	hexBytes, hexErr := hex.DecodeString(hexRepresentation)
	if hexErr != nil {
		return nil, hexErr
	}
	paddedBytes, padErr := padBytes(hexBytes, 32)
	if padErr != nil {
		return nil, padErr
	}
	return paddedBytes, nil
}

// To8PaddedHexBytes takes a number and returns its hex representation padded to 8 bytes.
func To8PaddedHexBytes(number uint64) ([]byte, error) {
	hexRepresentation := strconv.FormatUint(number, 16)
	// Make sure hex representation has even length.
	// The `strconv.FormatUint` can return odd length hex encodings.
	// For example, `strconv.FormatUint(10, 16)` returns `a`.
	// Thus, we need to pad it.
	if len(hexRepresentation)%2 == 1 {
		hexRepresentation = "0" + hexRepresentation
	}
	hexBytes, hexErr := hex.DecodeString(hexRepresentation)
	if hexErr != nil {
		return nil, hexErr
	}
	paddedBytes, padErr := padBytes(hexBytes, 8)
	if padErr != nil {
		return nil, padErr
	}
	return paddedBytes, nil
}
