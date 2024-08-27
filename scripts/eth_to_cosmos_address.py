import bech32


# This function replicates the Cosmos SDK ConvertAndEncode function:
# https://github.com/cosmos/cosmos-sdk/blob/v0.50.4/types/bech32/bech32.go#L10
def convert_and_encode(hrp: str, data: bytes):
    converted = bech32.convertbits(data, 8, 5, pad=True)
    return bech32.bech32_encode(hrp, converted)


# This is the Ethereum address we want to convert, and the expected mapping.
eth_address = "0x62d221db49aef5632f59b900b2ca90e52ecc0a80"
eth_address_bz = bytes.fromhex(eth_address[2:])
expected_cosmos_address = "fuelsequencer1vtfzrk6f4m6kxt6ehyqt9j5su5hvcz5q3dmlsm"
expected_valoper_address = "fuelsequencervaloper1vtfzrk6f4m6kxt6ehyqt9j5su5hvcz5qn0wwpn"

# Some constants that would be hard-coded.
BECH32_PREFIX = "fuelsequencer"
VALOPER_PREFIX = BECH32_PREFIX + "valoper"

# Generate address and convert to a Cosmos SDK account address.
cosmos_address = convert_and_encode(BECH32_PREFIX, eth_address_bz)
valoper_address = convert_and_encode(VALOPER_PREFIX, eth_address_bz)

print("ADDRESS: " + cosmos_address)
print("MATCH? : " + str(cosmos_address == expected_cosmos_address))
print()
print("VALOPER: " + valoper_address)
print("MATCH? : " + str(valoper_address == expected_valoper_address))
