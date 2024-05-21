import bech32


# This function replicates the Cosmos SDK ConvertAndEncode function:
# https://github.com/cosmos/cosmos-sdk/blob/v0.50.4/types/bech32/bech32.go#L10
def convert_and_encode(hrp: str, data: bytes):
    converted = bech32.convertbits(data, 8, 5, pad=True)
    return bech32.bech32_encode(hrp, converted)


# This is the Ethereum address we want to convert, and the expected mapping.
eth_address = "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266"
eth_address_bz = bytes.fromhex(eth_address[2:])
expected_cosmos_address = "fuelsequencer17w0adeg64ky0daxwd2ugyuneellmjgnx5dpmtz"

# Some constants that would be hard-coded.
BECH32_PREFIX = "fuelsequencer"

# Generate address and convert to a Cosmos SDK account address.
cosmos_address = convert_and_encode(BECH32_PREFIX, eth_address_bz)

print("ADDRESS: " + cosmos_address)
print("MATCH? : " + str(cosmos_address == expected_cosmos_address))
