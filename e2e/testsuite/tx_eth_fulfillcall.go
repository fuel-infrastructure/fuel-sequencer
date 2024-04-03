package testsuite

import "github.com/ethereum/go-ethereum/common"

const gatewayContractFulfillCallJSON = `
[
  {
	 "type":"function",
	 "name":"fulfillCall",
	 "inputs":[
		{
		   "name":"_functionId",
		   "type":"bytes32",
		   "internalType":"bytes32"
		},
		{
		   "name":"_input",
		   "type":"bytes",
		   "internalType":"bytes"
		},
		{
		   "name":"_output",
		   "type":"bytes",
		   "internalType":"bytes"
		},
		{
		   "name":"_proof",
		   "type":"bytes",
		   "internalType":"bytes"
		},
		{
		   "name":"_callbackAddress",
		   "type":"address",
		   "internalType":"address"
		},
		{
		   "name":"_callbackData",
		   "type":"bytes",
		   "internalType":"bytes"
		}
	 ],
	 "outputs":[
		
	 ],
	 "stateMutability":"nonpayable"
  }
]
`

func PackFulfillCall(
	functionId []byte,
	input []byte,
	output []byte,
	proof []byte,
	callbackAddress common.Address,
	callbackData []byte,
) []byte {
	return packCall(gatewayContractFulfillCallJSON, "fulfillCall", []interface{}{
		functionId,
		input,
		output,
		proof,
		callbackAddress,
		callbackData,
	})
}
