from typing import Dict, List


def get_update_consensus_module_params_proposal(
    deposit: str,
    block_max_bytes: str,
    block_max_gas: str,
    evidence_max_age_num_blocks: str,
    evidence_max_age_duration: str,
    evidence_max_bytes: str,
    validator_pub_key_types: List[str],
    proposal_title: str = "Proposal title",
    proposal_summary: str = "Proposal summary",
) -> Dict:
    return {
        "title": proposal_title,
        "summary": proposal_summary,
        "deposit": deposit,
        "messages": [
            {
                "@type": "/cosmos.consensus.v1.MsgUpdateParams",
                "authority": "fuelsequencer10d07y265gmmuvt4z0w9aw880jnsr700jdjfvk3",
                "block": {
                    "max_bytes": block_max_bytes,
                    "max_gas": block_max_gas,
                },
                "evidence": {
                    "max_age_num_blocks": evidence_max_age_num_blocks,
                    "max_age_duration": evidence_max_age_duration,
                    "max_bytes": evidence_max_bytes,
                },
                "validator": {
                    "pub_key_types": validator_pub_key_types,
                },
            }
        ]
    }
