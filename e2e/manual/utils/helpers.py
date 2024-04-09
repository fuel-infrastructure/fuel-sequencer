import json

from utils.classes import Chain


def get_balance_of(
        chain: Chain, address: str, denom: str
) -> int:
    # Get token balance.
    balances = json.loads(chain.query_balance_by_address(address))
    balances = balances['balances']
    filtered = [b for b in balances if b['denom'] == denom]

    if len(filtered) == 1:
        return int(filtered[0]['amount'])
    if len(filtered) > 1:
        raise Exception(f"Found multiple balances with same denom?? {filtered}")
    else:
        return 0
