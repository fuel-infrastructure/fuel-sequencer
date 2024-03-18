package testsuite

import "math/big"

const proxyContractDepositABIJSON = `
[
  {
    "type": "function",
    "name": "deposit",
    "inputs": [
      {
        "name": "_amount",
        "type": "uint256",
        "internalType": "uint256"
      },
      {
        "name": "_to",
        "type": "string",
        "internalType": "string"
      },
      {
        "name": "_duration",
        "type": "uint256",
        "internalType": "uint256"
      }
    ],
    "outputs": [],
    "stateMutability": "nonpayable"
  }
]
`

func PackDeposit(amount *big.Int, to string, duration *big.Int) []byte {
	return packCall(proxyContractDepositABIJSON, "deposit", []interface{}{
		amount,
		to,
		duration,
	})
}
