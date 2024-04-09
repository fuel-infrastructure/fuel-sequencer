from typing import Dict, List


def get_update_bridge_module_params_proposal(
    bridge_denom: str,
    ethereum_proxy_contract_address: str,
    authorize_messages_allowed: List[str],
    supply_delta_period: str,
    vesting_start_time: str,
    additional_blocked_addresses: List[str],
) -> Dict:
    return {
        "title": "Proposal title",
        "summary": "Proposal summary",
        "deposit": "10000000utest",
        "messages": [
            {
                "@type": "/fuelsequencer.bridge.MsgUpdateParams",
                "authority": "fuelsequencer10d07y265gmmuvt4z0w9aw880jnsr700jdjfvk3",
                "params": {
                    "bridge_denom": bridge_denom,
                    "ethereum_proxy_contract_address": ethereum_proxy_contract_address,
                    "authorize_messages_allowed": authorize_messages_allowed,
                    "supply_delta_period": supply_delta_period,
                    "vesting_start_time": vesting_start_time,
                    "additional_blocked_addresses": additional_blocked_addresses,
                }
            }
        ]
    }
