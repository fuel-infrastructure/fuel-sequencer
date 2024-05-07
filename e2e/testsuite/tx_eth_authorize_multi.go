package testsuite

import sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"

func PackAuthorizeMulti(bytes []byte) []byte {
	return packCall(
		sidecartypes.MockSequencerProxyContractABI,
		sidecartypes.MockAuthorizeMultiFunctionName,
		[]interface{}{
			bytes,
		},
	)
}
