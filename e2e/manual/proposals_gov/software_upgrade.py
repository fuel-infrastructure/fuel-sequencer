from typing import Dict


def get_software_upgrade_proposal(
        name: str, height: int, info: str,
) -> Dict:
    return {
        "title": "Upgrade to seq-testnet-1.3 (increase-power-reduction)",
        "summary": "This is a proposal to upgrade the network to seq-testnet-1.3 at block 100000",
        "deposit": "10000000000000000000000test",
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
        ],
        "expedited": True,
    }
