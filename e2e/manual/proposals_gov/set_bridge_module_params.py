from typing import Dict, List


def get_update_bridge_module_params_proposal(
    deposit: str,
    bridge_denom: str,
    bridge_denom_total_supply: str,
    ethereum_proxy_contract_address: str,
    authorize_messages_allowed: List[str],
    supply_delta_period: str,
    vesting_start_time: str,
    additional_blocked_addresses: List[str],
    max_eth_block_update_delay: str,
    injected_event_tx_max_bytes: str,
    sequencer_txs_allocation: str,
    max_authorize_messages: str,
    proposal_title: str = "Proposal title",
    proposal_summary: str = "Proposal summary",
) -> Dict:
    return {
        "title": proposal_title,
        "summary": proposal_summary,
        "deposit": deposit,
        "messages": [
            {
                "@type": "/fuelsequencer.bridge.v1.MsgUpdateParams",
                "authority": "fuelsequencer10d07y265gmmuvt4z0w9aw880jnsr700jdjfvk3",
                "params": {
                    "bridge_denom": bridge_denom,
                    "bridge_denom_total_supply": bridge_denom_total_supply,
                    "ethereum_proxy_contract_address": ethereum_proxy_contract_address,
                    "authorize_messages_allowed": authorize_messages_allowed,
                    "supply_delta_period": supply_delta_period,
                    "vesting_start_time": vesting_start_time,
                    "additional_blocked_addresses": additional_blocked_addresses,
                    "max_eth_block_update_delay": max_eth_block_update_delay,
                    "injected_event_tx_max_bytes": injected_event_tx_max_bytes,
                    "sequencer_txs_allocation": sequencer_txs_allocation,
                    "max_authorize_messages": max_authorize_messages,
                }
            }
        ]
    }
