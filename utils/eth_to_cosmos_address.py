import hashlib

import bech32


# This function replicates the Cosmos SDK Module function:
# https://github.com/cosmos/cosmos-sdk/blob/v0.50.4/types/address/hash.go#L73
def module(module_name: str, derivation_key: bytes) -> bytes:
    module_bz = module_name.encode('utf-8')
    padded_module_bz = bytearray(module_bz)
    padded_module_bz.extend(b'\x00')
    padded_module_bz_plus_derivation_key = padded_module_bz + derivation_key

    h = hashlib.new("sha256")
    h.update("module".encode('utf-8'))
    a_digest = h.digest()

    h = hashlib.new("sha256")
    h.update(a_digest)
    h.update(padded_module_bz_plus_derivation_key)
    ab_digest = h.digest()

    return ab_digest


# This function replicates the Cosmos SDK ConvertAndEncode function:
# https://github.com/cosmos/cosmos-sdk/blob/v0.50.4/types/bech32/bech32.go#L10
def convert_and_encode(hrp: str, data: bytes):
    converted = bech32.convertbits(data, 8, 5, pad=True)
    return bech32.bech32_encode(hrp, converted)


# This is the Ethereum address we want to convert, and the expected mapping.
eth_address = "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"
eth_address_bz = bytes.fromhex(eth_address[2:])
expected_cosmos_address = "fuelsequencer13tch2uhman7dhjjphmx9uwx7kvg2kqfj5y56hsmljlv93pgma5vqyks99k"

# Some constants that would be hard-coded.
MODULE_NAME = "bridge"
BECH32_PREFIX = "fuelsequencer"

# Generate address and convert to a Cosmos SDK account address.
address = module(MODULE_NAME, eth_address_bz)
cosmos_address = convert_and_encode(BECH32_PREFIX, address)

print("ADDRESS: " + cosmos_address)
print("MATCH? : " + str(cosmos_address == expected_cosmos_address))
