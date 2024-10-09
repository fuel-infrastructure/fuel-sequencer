from typing import Dict


def get_update_sequencing_module_params_proposal(
    deposit: str,
    max_blob_size_bytes: str,
    sequencer_tx_max_bytes: str,
    proposal_title: str = "Proposal title",
    proposal_summary: str = "Proposal summary",
) -> Dict:
    return {
        "title": proposal_title,
        "summary": proposal_summary,
        "deposit": deposit,
        "messages": [
            {
                "@type": "/fuelsequencer.sequencing.v1.MsgUpdateParams",
                "authority": "fuelsequencer10d07y265gmmuvt4z0w9aw880jnsr700jdjfvk3",
                "params": {
                    "max_blob_size_bytes": max_blob_size_bytes,
                    "sequencer_tx_max_bytes": sequencer_tx_max_bytes,
                }
            }
        ]
    }
