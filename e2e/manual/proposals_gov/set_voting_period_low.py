from typing import Dict


def get_set_voting_period_low_proposal() -> Dict:
    return {
        "title": "Update voting period to 10s",
        "description": "Update voting period to 10s for quicker proposals",
        "changes": [
            {
                "subspace": "gov",
                "key": "votingparams",
                "value": {
                    "voting_period": "10000000000"
                }
            }
        ],
        "deposit": "10000000utest"
    }
