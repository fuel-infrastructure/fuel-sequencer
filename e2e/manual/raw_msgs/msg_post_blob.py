from typing import Dict, List


def get_msg_post_blob(
        sender: str, topic: str, order: str, data: str, gas: str, fee: List,
) -> Dict:
    return \
        {
            "body": {
                "messages": [
                    {
                        "@type": "/fuelsequencer.sequencing.v1.MsgPostBlob",
                        "from": sender,
                        "topic": topic,
                        "order": order,
                        "data": data
                    }
                ],
                "memo": "",
                "timeout_height": "0",
                "extension_options": [],
                "non_critical_extension_options": []
            },
            "auth_info": {
                "signer_infos": [],
                "fee": {
                    "amount": fee,
                    "gas_limit": gas,
                    "payer": "",
                    "granter": ""
                }
            },
            "signatures": []
        }
