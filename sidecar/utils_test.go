package sidecar

import (
	"encoding/hex"
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

type HeadUpdateEvent struct {
	BlockNumber uint64      `json:"blockNumber"`
	HeaderHash  common.Hash `json:"headerHash"`
}

type DataCommitmentStoredEvent struct {
	ProofNonce *big.Int `json:"proofNonce"`
}

type CallEvent struct {
	InputHash  common.Hash `json:"inputHash"`
	OutputHash common.Hash `json:"outputHash"`
}

func TestStuff(t *testing.T) {

	const gatewayAbi = `[{"type":"function","name":"deployAndRegisterFunction","inputs":[{"name":"_owner","type":"address","internalType":"address"},{"name":"_bytecode","type":"bytes","internalType":"bytes"},{"name":"_salt","type":"bytes32","internalType":"bytes32"}],"outputs":[{"name":"functionId","type":"bytes32","internalType":"bytes32"},{"name":"verifier","type":"address","internalType":"address"}],"stateMutability":"nonpayable"},{"type":"function","name":"deployAndUpdateFunction","inputs":[{"name":"_bytecode","type":"bytes","internalType":"bytes"},{"name":"_salt","type":"bytes32","internalType":"bytes32"}],"outputs":[{"name":"functionId","type":"bytes32","internalType":"bytes32"},{"name":"verifier","type":"address","internalType":"address"}],"stateMutability":"nonpayable"},{"type":"function","name":"fulfillCall","inputs":[{"name":"_functionId","type":"bytes32","internalType":"bytes32"},{"name":"_input","type":"bytes","internalType":"bytes"},{"name":"_output","type":"bytes","internalType":"bytes"},{"name":"_proof","type":"bytes","internalType":"bytes"},{"name":"_callbackAddress","type":"address","internalType":"address"},{"name":"_callbackData","type":"bytes","internalType":"bytes"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"getFunctionId","inputs":[{"name":"_owner","type":"address","internalType":"address"},{"name":"_salt","type":"bytes32","internalType":"bytes32"}],"outputs":[{"name":"functionId","type":"bytes32","internalType":"bytes32"}],"stateMutability":"pure"},{"type":"function","name":"isCallback","inputs":[],"outputs":[{"name":"","type":"bool","internalType":"bool"}],"stateMutability":"view"},{"type":"function","name":"nonce","inputs":[],"outputs":[{"name":"","type":"uint32","internalType":"uint32"}],"stateMutability":"view"},{"type":"function","name":"outputs","inputs":[{"name":"","type":"bytes32","internalType":"bytes32"}],"outputs":[{"name":"","type":"bytes","internalType":"bytes"}],"stateMutability":"view"},{"type":"function","name":"registerFunction","inputs":[{"name":"_owner","type":"address","internalType":"address"},{"name":"_verifier","type":"address","internalType":"address"},{"name":"_salt","type":"bytes32","internalType":"bytes32"}],"outputs":[{"name":"functionId","type":"bytes32","internalType":"bytes32"}],"stateMutability":"nonpayable"},{"type":"function","name":"requestCall","inputs":[{"name":"_functionId","type":"bytes32","internalType":"bytes32"},{"name":"_input","type":"bytes","internalType":"bytes"},{"name":"_entryAddress","type":"address","internalType":"address"},{"name":"_entryCalldata","type":"bytes","internalType":"bytes"},{"name":"_entryGasLimit","type":"uint32","internalType":"uint32"}],"outputs":[],"stateMutability":"payable"},{"type":"function","name":"requestCall","inputs":[{"name":"_functionId","type":"bytes32","internalType":"bytes32"},{"name":"_input","type":"bytes","internalType":"bytes"},{"name":"_entryAddress","type":"address","internalType":"address"},{"name":"_entryCalldata","type":"bytes","internalType":"bytes"},{"name":"_entryGasLimit","type":"uint32","internalType":"uint32"},{"name":"","type":"address","internalType":"address"},{"name":"","type":"bytes","internalType":"bytes"}],"outputs":[],"stateMutability":"payable"},{"type":"function","name":"requestCallback","inputs":[{"name":"_functionId","type":"bytes32","internalType":"bytes32"},{"name":"_input","type":"bytes","internalType":"bytes"},{"name":"_context","type":"bytes","internalType":"bytes"},{"name":"_callbackSelector","type":"bytes4","internalType":"bytes4"},{"name":"_callbackGasLimit","type":"uint32","internalType":"uint32"}],"outputs":[{"name":"","type":"bytes32","internalType":"bytes32"}],"stateMutability":"payable"},{"type":"function","name":"requests","inputs":[{"name":"","type":"uint32","internalType":"uint32"}],"outputs":[{"name":"","type":"bytes32","internalType":"bytes32"}],"stateMutability":"view"},{"type":"function","name":"updateFunction","inputs":[{"name":"_verifier","type":"address","internalType":"address"},{"name":"_salt","type":"bytes32","internalType":"bytes32"}],"outputs":[{"name":"functionId","type":"bytes32","internalType":"bytes32"}],"stateMutability":"nonpayable"},{"type":"function","name":"verifiedCall","inputs":[{"name":"_functionId","type":"bytes32","internalType":"bytes32"},{"name":"_input","type":"bytes","internalType":"bytes"}],"outputs":[{"name":"","type":"bytes","internalType":"bytes"}],"stateMutability":"view"},{"type":"function","name":"verifiedFunctionId","inputs":[],"outputs":[{"name":"","type":"bytes32","internalType":"bytes32"}],"stateMutability":"view"},{"type":"function","name":"verifiedInputHash","inputs":[],"outputs":[{"name":"","type":"bytes32","internalType":"bytes32"}],"stateMutability":"view"},{"type":"function","name":"verifiedOutput","inputs":[],"outputs":[{"name":"","type":"bytes","internalType":"bytes"}],"stateMutability":"view"},{"type":"function","name":"verifierOwners","inputs":[{"name":"","type":"bytes32","internalType":"bytes32"}],"outputs":[{"name":"","type":"address","internalType":"address"}],"stateMutability":"view"},{"type":"function","name":"verifiers","inputs":[{"name":"","type":"bytes32","internalType":"bytes32"}],"outputs":[{"name":"","type":"address","internalType":"address"}],"stateMutability":"view"},{"type":"event","name":"Call","inputs":[{"name":"functionId","type":"bytes32","indexed":true,"internalType":"bytes32"},{"name":"inputHash","type":"bytes32","indexed":false,"internalType":"bytes32"},{"name":"outputHash","type":"bytes32","indexed":false,"internalType":"bytes32"}],"anonymous":false},{"type":"event","name":"Deployed","inputs":[{"name":"bytecodeHash","type":"bytes32","indexed":true,"internalType":"bytes32"},{"name":"salt","type":"bytes32","indexed":true,"internalType":"bytes32"},{"name":"deployedAddress","type":"address","indexed":true,"internalType":"address"}],"anonymous":false},{"type":"event","name":"FunctionRegistered","inputs":[{"name":"functionId","type":"bytes32","indexed":true,"internalType":"bytes32"},{"name":"verifier","type":"address","indexed":false,"internalType":"address"},{"name":"salt","type":"bytes32","indexed":false,"internalType":"bytes32"},{"name":"owner","type":"address","indexed":false,"internalType":"address"}],"anonymous":false},{"type":"event","name":"FunctionVerifierUpdated","inputs":[{"name":"functionId","type":"bytes32","indexed":true,"internalType":"bytes32"},{"name":"verifier","type":"address","indexed":false,"internalType":"address"}],"anonymous":false},{"type":"event","name":"ProverUpdated","inputs":[{"name":"functionId","type":"bytes32","indexed":true,"internalType":"bytes32"},{"name":"prover","type":"address","indexed":true,"internalType":"address"},{"name":"added","type":"bool","indexed":false,"internalType":"bool"}],"anonymous":false},{"type":"event","name":"RequestCall","inputs":[{"name":"functionId","type":"bytes32","indexed":true,"internalType":"bytes32"},{"name":"input","type":"bytes","indexed":false,"internalType":"bytes"},{"name":"entryAddress","type":"address","indexed":false,"internalType":"address"},{"name":"entryCalldata","type":"bytes","indexed":false,"internalType":"bytes"},{"name":"entryGasLimit","type":"uint32","indexed":false,"internalType":"uint32"},{"name":"sender","type":"address","indexed":false,"internalType":"address"},{"name":"feeAmount","type":"uint256","indexed":false,"internalType":"uint256"}],"anonymous":false},{"type":"event","name":"RequestCallback","inputs":[{"name":"nonce","type":"uint32","indexed":true,"internalType":"uint32"},{"name":"functionId","type":"bytes32","indexed":true,"internalType":"bytes32"},{"name":"input","type":"bytes","indexed":false,"internalType":"bytes"},{"name":"context","type":"bytes","indexed":false,"internalType":"bytes"},{"name":"callbackAddress","type":"address","indexed":false,"internalType":"address"},{"name":"callbackSelector","type":"bytes4","indexed":false,"internalType":"bytes4"},{"name":"callbackGasLimit","type":"uint32","indexed":false,"internalType":"uint32"},{"name":"feeAmount","type":"uint256","indexed":false,"internalType":"uint256"}],"anonymous":false},{"type":"event","name":"RequestFulfilled","inputs":[{"name":"nonce","type":"uint32","indexed":true,"internalType":"uint32"},{"name":"functionId","type":"bytes32","indexed":true,"internalType":"bytes32"},{"name":"inputHash","type":"bytes32","indexed":false,"internalType":"bytes32"},{"name":"outputHash","type":"bytes32","indexed":false,"internalType":"bytes32"}],"anonymous":false},{"type":"event","name":"SetFeeVault","inputs":[{"name":"oldFeeVault","type":"address","indexed":true,"internalType":"address"},{"name":"newFeeVault","type":"address","indexed":true,"internalType":"address"}],"anonymous":false},{"type":"event","name":"WhitelistStatusUpdated","inputs":[{"name":"functionId","type":"bytes32","indexed":true,"internalType":"bytes32"},{"name":"status","type":"uint8","indexed":false,"internalType":"enum WhitelistStatus"}],"anonymous":false},{"type":"error","name":"CallFailed","inputs":[{"name":"callbackAddress","type":"address","internalType":"address"},{"name":"callbackData","type":"bytes","internalType":"bytes"}]},{"type":"error","name":"CallbackFailed","inputs":[{"name":"callbackSelector","type":"bytes4","internalType":"bytes4"},{"name":"output","type":"bytes","internalType":"bytes"},{"name":"context","type":"bytes","internalType":"bytes"}]},{"type":"error","name":"EmptyBytecode","inputs":[]},{"type":"error","name":"FailedDeploy","inputs":[]},{"type":"error","name":"FunctionAlreadyRegistered","inputs":[{"name":"functionId","type":"bytes32","internalType":"bytes32"}]},{"type":"error","name":"InvalidCall","inputs":[{"name":"functionId","type":"bytes32","internalType":"bytes32"},{"name":"input","type":"bytes","internalType":"bytes"}]},{"type":"error","name":"InvalidProof","inputs":[{"name":"verifier","type":"address","internalType":"address"},{"name":"inputHash","type":"bytes32","internalType":"bytes32"},{"name":"outputHash","type":"bytes32","internalType":"bytes32"},{"name":"proof","type":"bytes","internalType":"bytes"}]},{"type":"error","name":"InvalidRequest","inputs":[{"name":"nonce","type":"uint32","internalType":"uint32"},{"name":"expectedRequestHash","type":"bytes32","internalType":"bytes32"},{"name":"requestHash","type":"bytes32","internalType":"bytes32"}]},{"type":"error","name":"NotFunctionOwner","inputs":[{"name":"owner","type":"address","internalType":"address"},{"name":"actualOwner","type":"address","internalType":"address"}]},{"type":"error","name":"OnlyProver","inputs":[{"name":"functionId","type":"bytes32","internalType":"bytes32"},{"name":"sender","type":"address","internalType":"address"}]},{"type":"error","name":"RecoverFailed","inputs":[]},{"type":"error","name":"ReentrantFulfill","inputs":[]},{"type":"error","name":"VerifierAlreadyUpdated","inputs":[{"name":"functionId","type":"bytes32","internalType":"bytes32"}]},{"type":"error","name":"VerifierCannotBeZero","inputs":[]}]`
	const fuelStreamXAbi = `[{"type":"constructor","inputs":[{"name":"_params","type":"tuple","internalType":"structFuelStreamX.InitParameters","components":[{"name":"guardian","type":"address","internalType":"address"},{"name":"gateway","type":"address","internalType":"address"},{"name":"height","type":"uint64","internalType":"uint64"},{"name":"header","type":"bytes32","internalType":"bytes32"},{"name":"nextHeaderFunctionId","type":"bytes32","internalType":"bytes32"},{"name":"headerRangeFunctionId","type":"bytes32","internalType":"bytes32"}]}],"stateMutability":"nonpayable"},{"type":"function","name":"Authorize","inputs":[{"name":"_message","type":"bytes","internalType":"bytes"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"BRIDGE_COMMITMENT_MAX","inputs":[],"outputs":[{"name":"","type":"uint64","internalType":"uint64"}],"stateMutability":"view"},{"type":"function","name":"VERSION","inputs":[],"outputs":[{"name":"","type":"string","internalType":"string"}],"stateMutability":"pure"},{"type":"function","name":"blockHeightToHeaderHash","inputs":[{"name":"","type":"uint64","internalType":"uint64"}],"outputs":[{"name":"","type":"bytes32","internalType":"bytes32"}],"stateMutability":"view"},{"type":"function","name":"commitHeaderRange","inputs":[{"name":"_targetBlock","type":"uint64","internalType":"uint64"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"commitNextHeader","inputs":[{"name":"_trustedBlock","type":"uint64","internalType":"uint64"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"deposit","inputs":[{"name":"_amount","type":"uint256","internalType":"uint256"},{"name":"_to","type":"string","internalType":"string"},{"name":"_duration","type":"uint256","internalType":"uint256"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"frozen","inputs":[],"outputs":[{"name":"","type":"bool","internalType":"bool"}],"stateMutability":"view"},{"type":"function","name":"gateway","inputs":[],"outputs":[{"name":"","type":"address","internalType":"address"}],"stateMutability":"view"},{"type":"function","name":"headerRangeFunctionId","inputs":[],"outputs":[{"name":"","type":"bytes32","internalType":"bytes32"}],"stateMutability":"view"},{"type":"function","name":"latestBlock","inputs":[],"outputs":[{"name":"","type":"uint64","internalType":"uint64"}],"stateMutability":"view"},{"type":"function","name":"nextHeaderFunctionId","inputs":[],"outputs":[{"name":"","type":"bytes32","internalType":"bytes32"}],"stateMutability":"view"},{"type":"function","name":"processSequencerWithdrawalMessage","inputs":[{"name":"_proofNonce","type":"uint256","internalType":"uint256"},{"name":"bridgeCommitmentLeaf","type":"tuple","internalType":"structFuelStreamX.BridgeCommitmentLeaf","components":[{"name":"height","type":"uint256","internalType":"uint256"},{"name":"dataHash","type":"bytes32","internalType":"bytes32"},{"name":"resultsHash","type":"bytes32","internalType":"bytes32"}]},{"name":"bridgeCommitmentLeafProof","type":"tuple","internalType":"structBinaryMerkleProof","components":[{"name":"sideNodes","type":"bytes32[]","internalType":"bytes32[]"},{"name":"key","type":"uint256","internalType":"uint256"},{"name":"numLeaves","type":"uint256","internalType":"uint256"}]},{"name":"txResultMarshalled","type":"bytes","internalType":"bytes"},{"name":"txResultProof","type":"tuple","internalType":"structBinaryMerkleProof","components":[{"name":"sideNodes","type":"bytes32[]","internalType":"bytes32[]"},{"name":"key","type":"uint256","internalType":"uint256"},{"name":"numLeaves","type":"uint256","internalType":"uint256"}]}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"requestHeaderRange","inputs":[{"name":"_targetBlock","type":"uint64","internalType":"uint64"}],"outputs":[],"stateMutability":"payable"},{"type":"function","name":"requestNextHeader","inputs":[],"outputs":[],"stateMutability":"payable"},{"type":"function","name":"setBridgeCommitmentRoot","inputs":[{"name":"nonce","type":"uint256","internalType":"uint256"},{"name":"bridgeCommitmentRoot","type":"bytes32","internalType":"bytes32"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"state_dataCommitments","inputs":[{"name":"","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"","type":"bytes32","internalType":"bytes32"}],"stateMutability":"view"},{"type":"function","name":"state_proofNonce","inputs":[],"outputs":[{"name":"","type":"uint256","internalType":"uint256"}],"stateMutability":"view"},{"type":"function","name":"updateFreeze","inputs":[{"name":"_freeze","type":"bool","internalType":"bool"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"updateFunctionIds","inputs":[{"name":"_headerRangeFunctionId","type":"bytes32","internalType":"bytes32"},{"name":"_nextHeaderFunctionId","type":"bytes32","internalType":"bytes32"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"updateGateway","inputs":[{"name":"_gateway","type":"address","internalType":"address"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"updateGenesisState","inputs":[{"name":"_height","type":"uint32","internalType":"uint32"},{"name":"_header","type":"bytes32","internalType":"bytes32"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"withdrawalNoncesExecuted","inputs":[{"name":"","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"","type":"bool","internalType":"bool"}],"stateMutability":"view"},{"type":"event","name":"AuthorizeEvent","inputs":[{"name":"_from","type":"address","indexed":true,"internalType":"address"},{"name":"_message","type":"bytes","indexed":false,"internalType":"bytes"}],"anonymous":false},{"type":"event","name":"DataCommitmentStored","inputs":[{"name":"proofNonce","type":"uint256","indexed":false,"internalType":"uint256"},{"name":"startBlock","type":"uint64","indexed":true,"internalType":"uint64"},{"name":"endBlock","type":"uint64","indexed":true,"internalType":"uint64"},{"name":"dataCommitment","type":"bytes32","indexed":true,"internalType":"bytes32"}],"anonymous":false},{"type":"event","name":"HeadUpdate","inputs":[{"name":"blockNumber","type":"uint64","indexed":false,"internalType":"uint64"},{"name":"headerHash","type":"bytes32","indexed":false,"internalType":"bytes32"}],"anonymous":false},{"type":"event","name":"HeaderRangeRequested","inputs":[{"name":"trustedBlock","type":"uint64","indexed":true,"internalType":"uint64"},{"name":"trustedHeader","type":"bytes32","indexed":true,"internalType":"bytes32"},{"name":"targetBlock","type":"uint64","indexed":true,"internalType":"uint64"}],"anonymous":false},{"type":"event","name":"NextHeaderRequested","inputs":[{"name":"trustedBlock","type":"uint64","indexed":true,"internalType":"uint64"},{"name":"trustedHeader","type":"bytes32","indexed":true,"internalType":"bytes32"}],"anonymous":false},{"type":"event","name":"SendToSequencerEvent","inputs":[{"name":"_from","type":"address","indexed":true,"internalType":"address"},{"name":"_amount","type":"uint256","indexed":false,"internalType":"uint256"},{"name":"_to","type":"string","indexed":false,"internalType":"string"},{"name":"_duration","type":"uint256","indexed":false,"internalType":"uint256"}],"anonymous":false},{"type":"error","name":"ContractFrozen","inputs":[]},{"type":"error","name":"DataCommitmentNotFound","inputs":[]},{"type":"error","name":"LatestHeaderNotFound","inputs":[]},{"type":"error","name":"TargetBlockNotInRange","inputs":[]},{"type":"error","name":"TrustedBlockMismatch","inputs":[]},{"type":"error","name":"TrustedHeaderNotFound","inputs":[]}]`

	const log0Topic0 = "0x292f5abc3167175400fca463fa99530cda826ec53ec5eb1f3a2776006dacd75d"
	const log0Data = "0x000000000000000000000000000000000000000000000000000000000000000f000000000000000000000000000000000000000000000000000000000000000f"

	const log1Topic0 = "0x34dd3689f5bd77a60a3ff2e09483dcab032fa2f1fd7227af3e24bed21beab1cb"
	const log1Topic1 = "0x0000000000000000000000000000000000000000000000000000000000000001"
	const log1Topic2 = "0x000000000000000000000000000000000000000000000000000000000000000f"
	const log1Topic3 = "0xb0c735c140c95e0cd16631c0733c8f805b918b0530a1d678c3a527957ae877b5"
	const log1Data = "0x0000000000000000000000000000000000000000000000000000000000000001"

	const log2Topic0 = "0x41d7122d18af9f0c92f23bcea9d5fa416cadcd1ed2fc8e544a3c89b841ecfd15"
	const log2Topic1 = "0xa3c1274aadd82e4d12c8004c33fb244ca686dad4fcc8957fc5668588c11d9502"
	const log2Data = "0x3b46be95aec0e95a4bfa506307dfd5fd6e180f5ca786d42ca04beccf5ade94eda9f5f462515513ffc052fa2b84bd0c1d22b8fe8039480e18c08f040c3d10727e"

	fuelstreamxABI, err := abi.JSON(strings.NewReader(fuelStreamXAbi))
	require.NoError(t, err)
	gatewayABI, err := abi.JSON(strings.NewReader(gatewayAbi))
	require.NoError(t, err)

	// ---------------------------- Check event 1

	expectedHeadUpdateEventHashFn := crypto.Keccak256Hash([]byte("HeadUpdate(uint64,bytes32)")).Hex()
	require.Equal(t, expectedHeadUpdateEventHashFn, log0Topic0)
	// TODO: confirm the expected values are correct

	var event1 HeadUpdateEvent
	data, err := hex.DecodeString(log0Data[2:])
	require.NoError(t, err)
	err = fuelstreamxABI.UnpackIntoInterface(&event1, "HeadUpdate", data)
	require.NoError(t, err)

	expectedBlockNumber := 15
	expectedHeaderHash := "0x000000000000000000000000000000000000000000000000000000000000000f"
	require.EqualValues(t, expectedBlockNumber, event1.BlockNumber)
	require.EqualValues(t, expectedHeaderHash, event1.HeaderHash.Hex())
	// TODO: confirm the expected values are correct

	// ---------------------------- Check event 2

	expectedDataCommitmentStoredEventHashFn := crypto.Keccak256Hash(
		[]byte("DataCommitmentStored(uint256,uint64,uint64,bytes32)"),
	).Hex()
	expectedStartBlock := 1
	expectedEndBlock := 15
	expectedDataCommitment := "0xb0c735c140c95e0cd16631c0733c8f805b918b0530a1d678c3a527957ae877b5"
	// TODO: confirm the expected values are correct

	actualStartBlock, err := strconv.ParseUint(log1Topic1[2:], 16, 64)
	require.NoError(t, err)
	actualEndBlock, err := strconv.ParseUint(log1Topic2[2:], 16, 64)
	require.NoError(t, err)

	require.Equal(t, expectedDataCommitmentStoredEventHashFn, log1Topic0)
	require.EqualValues(t, expectedStartBlock, actualStartBlock)
	require.EqualValues(t, expectedEndBlock, actualEndBlock)
	require.Equal(t, expectedDataCommitment, log1Topic3)

	var event2 DataCommitmentStoredEvent
	data, err = hex.DecodeString(log1Data[2:])
	require.NoError(t, err)
	err = fuelstreamxABI.UnpackIntoInterface(&event2, "DataCommitmentStored", data)
	require.NoError(t, err)

	expectedProofNonce := 1
	require.EqualValues(t, expectedProofNonce, event2.ProofNonce.Uint64())
	// TODO: confirm the expected values are correct

	// ---------------------------- Check event 3

	expectedCallEventHashFn := crypto.Keccak256Hash([]byte("Call(bytes32,bytes32,bytes32)")).Hex()
	expectedHeaderRangeFunctionId := "0xa3c1274aadd82e4d12c8004c33fb244ca686dad4fcc8957fc5668588c11d9502"
	require.Equal(t, expectedCallEventHashFn, log2Topic0)
	require.Equal(t, expectedHeaderRangeFunctionId, log2Topic1)
	// TODO: confirm the expected values are correct

	var event3 CallEvent
	data, err = hex.DecodeString(log2Data[2:])
	require.NoError(t, err)
	err = gatewayABI.UnpackIntoInterface(&event3, "Call", data)
	require.NoError(t, err)

	expectedInputHash := "0x3b46be95aec0e95a4bfa506307dfd5fd6e180f5ca786d42ca04beccf5ade94ed"
	expectedOutputHash := "0xa9f5f462515513ffc052fa2b84bd0c1d22b8fe8039480e18c08f040c3d10727e"
	require.Equal(t, expectedInputHash, event3.InputHash.Hex())
	require.Equal(t, expectedOutputHash, event3.OutputHash.Hex())
	// TODO: confirm the expected values are correct
}
