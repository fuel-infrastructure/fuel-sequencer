from typing import Dict, List


def get_grant_authorisation_proposal(
        grantee: str, msg_type_urls: List[str]
) -> Dict:
    return {
        "title": "Proposal title",
        "summary": "Proposal summary",
        "deposit": "10000000utest",
        "messages": [
            {
                "@type": "/cosmos.authz.v1beta1.MsgGrant",
                "granter": "fuelsequencer10d07y265gmmuvt4z0w9aw880jnsr700jdjfvk3",
                "grantee": grantee,
                "grant": {
                    "authorization": {
                        "@type": "/cosmos.authz.v1beta1.GenericAuthorization",
                        "msg": msg_type_url,
                    },
                    # "expiration": {},  # no expiration if not specified
                },
            }
            for msg_type_url in msg_type_urls
        ],
    }
