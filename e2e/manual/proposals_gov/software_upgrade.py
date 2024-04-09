from typing import Dict


def get_software_upgrade_proposal(
        name: str, height: int, info: str,
) -> Dict:
    return {
        "title": "Proposal title",
        "summary": "Proposal summary",
        "deposit": "10000000utest",
        "messages": [
            {
                "@type": "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
                "authority": "fuelsequencer10d07y265gmmuvt4z0w9aw880jnsr700jdjfvk3",
                "plan": {
                    "name": name,
                    "height": height,
                    "info": info,
                }
            }
        ]
    }
