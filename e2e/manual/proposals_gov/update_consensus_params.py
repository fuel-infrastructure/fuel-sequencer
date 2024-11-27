from typing import Dict


# Note: the current values can be obtained from node:26657/consensus_params
#
# block must have:
# - max_bytes (e.g. "1048576")
# - max_gas (e.g. "-1")
#
# evidence must have:
# - max_age_num_blocks (e.g. "100000")
# - max_age_duration (e.g. "172800000000000ns")
# - max_bytes (e.g. "1048576")
#
# validator must have:
# - pub_key_types (e.g. ["ed25519"])
def get_update_consensus_params_proposal(
        block: Dict[str, str],
        evidence: Dict[str, str],
        validator: Dict[str, str],
) -> Dict:
    return {
        "title": "Proposal title",
        "summary": "Proposal summary",
        "deposit": "10000000utest",
        "messages": [
            {
                "@type": "/cosmos.consensus.v1.MsgUpdateParams",
                "authority": "fuelsequencer10d07y265gmmuvt4z0w9aw880jnsr700jdjfvk3",
                "block": block,
                "evidence": evidence,
                "validator": validator,
            },
        ]
    }
