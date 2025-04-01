import binascii
import hashlib

import requests

COMMIT_URL = "http://localhost:26657/commit?height={}"
VERIFICATION_URL = "http://localhost:1317/fuelsequencer/commitments/v1/bridge_commitment?start={}&end={}"
FROM = 1  # included
TO = 5  # excluded


def bridge_commitment_from_verification_url(from_height: int, to_height: int):
    data = requests.get(VERIFICATION_URL.format(from_height, to_height)).json()
    return data["bridge_commitment"]


def bridge_commitment_leaves(from_height: int, to_height: int):
    lrhs = []
    for i in range(from_height, to_height):
        commit = requests.get(COMMIT_URL.format(i)).json()
        lrh = bytes.fromhex(commit["result"]["signed_header"]["header"]["last_results_hash"])
        lrhs += [(i, lrh)]

    return lrhs

def to_32_padded_hex_bytes(number: int) -> bytes:
    # Convert the number to a hexadecimal string.
    hex_representation = format(number, 'x')

    # Ensure the hex representation has an even length.
    if len(hex_representation) % 2 == 1:
        hex_representation = '0' + hex_representation

    # Convert the hexadecimal string to bytes.
    try:
        hex_bytes = binascii.unhexlify(hex_representation)
    except binascii.Error as e:
        raise ValueError(f"Error decoding hexadecimal string: {e}")

    # Pad the bytes to ensure it is 32 bytes long.
    padded_bytes = hex_bytes.rjust(32, b'\x00')

    if len(padded_bytes) != 32:
        raise ValueError(
            "Padding failed: Resulting byte array is not 32 bytes.")

    return padded_bytes


def abi_encode_bridge_commitment_leaves(leaves):
    encoded_leaves = []

    for leaf in leaves:
        height, last_results_hash = leaf

        # Pad the height to 32 bytes.
        try:
            padded_height = to_32_padded_hex_bytes(height)
        except ValueError as e:
            raise ValueError(f"Error encoding leaf with height {height}: {e}")

        # Concatenate padded height and last_results_hash.
        encoded_leaf = padded_height + last_results_hash

        # Append the encoded leaf to the list.
        encoded_leaves.append(encoded_leaf)

    return encoded_leaves


def hash_from_byte_slices(items):
    return _hash_from_byte_slices(hashlib.sha256(), items)


def _hash_from_byte_slices(sha, items):
    if len(items) == 0:
        return empty_hash()
    elif len(items) == 1:
        return leaf_hash_opt(sha, items[0])
    else:
        k = get_split_point(len(items))
        left = _hash_from_byte_slices(sha, items[:k])
        right = _hash_from_byte_slices(sha, items[k:])
        return inner_hash_opt(sha, left, right)


def get_split_point(length):
    if length < 1:
        raise ValueError("Trying to split a tree with size < 1")
    bitlen = length.bit_length()
    k = 1 << (bitlen - 1)
    if k == length:
        k >>= 1
    return k


def empty_hash():
    return hashlib.sha256(b'').digest()


def leaf_hash_opt(sha, leaf):
    sha = sha.copy()
    sha.update(b'\x00' + leaf)
    return sha.digest()


def inner_hash_opt(sha, left, right):
    sha = sha.copy()
    sha.update(b'\x01' + left + right)
    return sha.digest()


bc_leaves = bridge_commitment_leaves(FROM, TO)
encoded = abi_encode_bridge_commitment_leaves(bc_leaves)
root = hash_from_byte_slices(encoded).hex().upper()
print(root)

verification = bridge_commitment_from_verification_url(FROM, TO)
print(verification)
