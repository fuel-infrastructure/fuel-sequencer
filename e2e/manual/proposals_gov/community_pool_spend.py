from typing import Dict, List


def get_community_pool_spend_proposal(
        recipient: str, amounts: List[Dict[str, str]],
) -> Dict:
    return {
        "title": "Proposal title",
        "summary": "Proposal summary",
        "deposit": "10000000utest",
        "messages": [
            {
                "@type": "/cosmos.distribution.v1beta1.MsgCommunityPoolSpend",
                "authority": "fuelsequencer10d07y265gmmuvt4z0w9aw880jnsr700jdjfvk3",
                "recipient": recipient,
                "amount": amounts,
            }
        ]
    }
