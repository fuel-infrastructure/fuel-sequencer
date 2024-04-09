from typing import Dict


def get_cancel_software_upgrade_proposal() -> Dict:
    return {
        "title": "Proposal title",
        "summary": "Proposal summary",
        "deposit": "10000000utest",
        "messages": [
            {
                "@type": "/cosmos.upgrade.v1beta1.MsgCancelUpgrade",
                "authority": "fuelsequencer10d07y265gmmuvt4z0w9aw880jnsr700jdjfvk3",
            }
        ]
    }
