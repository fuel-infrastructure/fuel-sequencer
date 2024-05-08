package testsuite

import sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"

func PackMockAuthorize(bytes []byte) []byte {
	return packCall(
		sidecartypes.MockSequencerProxyContractABI,
		sidecartypes.MockAuthorizeFunctionName,
		[]interface{}{
			bytes,
		},
	)
}
