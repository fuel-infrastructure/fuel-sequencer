package testsuite

const proxyContractAuthorizeMultiABIJSON = `
[
  {
    "type": "function",
    "name": "AuthorizeMulti",
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

func PackAuthorizeMulti(bytes []byte) []byte {
	return packCall(proxyContractAuthorizeMultiABIJSON, "AuthorizeMulti", []interface{}{
		bytes,
	})
}
