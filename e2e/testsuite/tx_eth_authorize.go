package testsuite

const proxyContractAuthorizeABIJSON = `
[
  {
    "type": "function",
    "name": "Authorize",
    "inputs": [
      {
        "name": "_message",
        "type": "bytes",
        "internalType": "bytes"
      }
    ],
    "outputs": [],
    "stateMutability": "nonpayable"
  }
]
`

func PackAuthorize(bytes []byte) []byte {
	return packCall(proxyContractAuthorizeABIJSON, "Authorize", []interface{}{
		bytes,
	})
}
