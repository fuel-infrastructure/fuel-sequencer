from typing import Tuple

import bech32


# This function replicates the Cosmos SDK DecodeAndConvert function:
# https://github.com/cosmos/cosmos-sdk/blob/v0.50.4/types/bech32/bech32.go#L20
def decode_and_convert(bech: str) -> Tuple[str, bytes]:
    hrp, data = bech32.bech32_decode(bech)
    converted = bech32.convertbits(data, 5, 8, pad=False)
    return hrp, bytes(converted)


# This is the Cosmos account or valoper address in bech32
# format that we want to convert, and the expected mapping.
bech32_address = "fuelsequencervaloper1vtfzrk6f4m6kxt6ehyqt9j5su5hvcz5qn0wwpn"
expected_eth_address = "0x62d221db49aef5632f59b900b2ca90e52ecc0a80"

# Generate address and convert to an Ethereum address.
_, bech32_address_bz = decode_and_convert(bech32_address)
eth_address = "0x" + bech32_address_bz.hex()

print("ADDRESS: " + eth_address)
print("MATCH? : " + str(eth_address == expected_eth_address))
