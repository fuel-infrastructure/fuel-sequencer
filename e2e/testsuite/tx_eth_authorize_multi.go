package testsuite

import sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"

func PackMockAuthorizeMulti(bytes []byte) []byte {
	return packCall(
		sidecartypes.MockSequencerProxyContractABI,
		sidecartypes.MockAuthorizeMultiFunctionName,
		[]interface{}{
			bytes,
		},
	)
}
