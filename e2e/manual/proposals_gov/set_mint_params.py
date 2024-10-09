from typing import Dict


def get_update_mint_module_params_proposal(
        deposit: str,
        mint_denom: str,
        inflation_rate_change: str,
        inflation_max: str,
        inflation_min: str,
        goal_bonded: str,
        blocks_per_year: str,
        proposal_title: str = "Proposal title",
        proposal_summary: str = "Proposal summary",
        expedited: bool = False,
) -> Dict:
    return {
        "title": proposal_title,
        "summary": proposal_summary,
        "deposit": deposit,
        "messages": [
            {
                "@type": "/cosmos.mint.v1beta1.MsgUpdateParams",
                "authority": "fuelsequencer10d07y265gmmuvt4z0w9aw880jnsr700jdjfvk3",
                "params": {
                    "mint_denom": mint_denom,
                    "inflation_rate_change": inflation_rate_change,
                    "inflation_max": inflation_max,
                    "inflation_min": inflation_min,
                    "goal_bonded": goal_bonded,
                    "blocks_per_year": blocks_per_year,
                }
            }
        ],
        "expedited": expedited,
    }
