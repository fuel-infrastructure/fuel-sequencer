package testsuite

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
)

func PackMint(address common.Address, amount *big.Int) []byte {
	return packCall(
		sidecartypes.TokenContractABI,
		sidecartypes.MintFunctionName,
		[]interface{}{
			address,
			amount,
		},
	)
}

func PackTransferAndCall(amount *big.Int) []byte {
	return packCall(
		sidecartypes.TokenContractABI,
		sidecartypes.TransferAndCallFunctionName,
		[]interface{}{
			common.HexToAddress(SEQUENCER_INTERFACE_CONTRACT),
			amount,
		},
	)
}

func PackAuthorize(data []byte) []byte {
	return PackBatchAuthorize([][]byte{data})
}

func PackBatchAuthorize(data [][]byte) []byte {
	return packCall(
		sidecartypes.SequencerInterfaceContractABI,
		sidecartypes.BatchAuthorizeFunctionName,
		[]interface{}{
			data,
		},
	)
}
