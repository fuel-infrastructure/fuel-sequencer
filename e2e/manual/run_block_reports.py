import time
from typing import Optional, Union

import requests
from dateutil import parser

from utils.networks import NetworkConfig, Networks

NETWORK = Networks.TESTNET  # Change me to load test other networks!
CONFIG = NetworkConfig(NETWORK)
print(f"Running block reports on {NETWORK}")

sequencer_rpc = CONFIG.seq_rpc
sequencer_rest = CONFIG.seq_rest
commit_query = "/commit?height="
block_query = "/block?height="
account_query = "/cosmos/auth/v1beta1/account_info/"


def get_account(address: str, height: Optional[Union[str, int]] = None):
    headers = {
        "x-cosmos-block-height": str(height)
    } if height is not None else {}
    return requests.get(
        f"{sequencer_rest}{account_query}{address}",
        headers=headers,
    ).json()


def get_block(n):
    return requests.get(f"{sequencer_rpc}{block_query}{n}").json()['result']


def get_commit(n):
    return requests.get(f"{sequencer_rpc}{commit_query}{n}").json()['result']


def get_latest_commit():
    return get_commit("")


def get_latest_block():
    return get_block("")


# Calculate block range based on latest height
latest_block = get_latest_commit()
latest_height = int(latest_block['signed_header']['header']['height'])
end_heights = [latest_height]
start_heights = [latest_height - 200]

# Optional: manual override of heights
# start_heights = []
# end_heights = []

# Get number of transactions in each block range for these signers
signers = [
    # "fuelsequencer1vtfzrk6f4m6kxt6ehyqt9j5su5hvcz5q3dmlsm",  # alice
    # "fuelsequencer1ptrdx8rzykw57suy540tlhfhxfclmykywhtycf",  # temp1
    # "fuelsequencer1z6mj9pgyd0gdt7zpnn5k5xgw3u99ammylkh0tn",  # temp2
    # "fuelsequencer18q22y7cyxsev5ttqkvqt6dse0zd6zejyqja2h0",  # temp3
    # "fuelsequencer1wnfm34d2wees97cl34qcev8de3ja2n8whh732u",  # temp4
    # "fuelsequencer1h4dr6td79wagkgt5z8qhzaazh30nnavgq8g7wf",  # temp5
    # "fuelsequencer1w683zjxx9pvnceakafaf3penc060uhggdqjsw3",  # temp6
    # "fuelsequencer16eqy4fjtrr5rf8e0gfsc48qlzdpfl5enng849y",  # temp7
    # "fuelsequencer1ngsagnggmmumc62220n6waav5wzjgm8hf8kzh0",  # temp8
    # "fuelsequencer1cyvqtp695llvpcu08xg5ndjsqre0nsash972vz",  # temp9
    # "fuelsequencer1lut8dr0473pxm9ayynay7hhya9n8dnsa6egh5p",  # temp10
]


def base64_decoded_size(encoded_str):
    l = len(encoded_str)
    p = encoded_str.count('=')  # Count padding characters
    return (l * 3) // 4 - p


def run_report_1(start_height: int, end_height: int):
    start_block = get_commit(start_height)
    start_block_timestamp = parser.parse(
        start_block['signed_header']['header']['time']
    )
    end_block = get_commit(end_height)
    end_block_timestamp = parser.parse(
        end_block['signed_header']['header']['time']
    )
    block_range_time = (end_block_timestamp - start_block_timestamp)
    seconds_per_block = block_range_time / (end_height - start_height)
    print(
        f"From {start_height} to {end_height}:\n"
        f"- Number of blocks: {end_height - start_height}\n"
        f"- Time elapsed: {block_range_time}\n"
        f"- Seconds per block: {seconds_per_block}"
    )


def run_report_2(start_height: int, end_height: int, address: str):
    sequence1 = int(get_account(address, start_height)['info']['sequence'])
    sequence2 = int(get_account(address, end_height)['info']['sequence'])
    print(
        f"Transactions from {start_height} to {end_height} ({address}): "
        f"{sequence2 - sequence1}"
    )


def run_report_3(start_height: int, end_height: int):
    previous_block_timestamp = None
    for n in range(start_height, end_height):

        try:
            block = get_block(n)
            block_timestamp = parser.parse(block['block']['header']['time'])

            block_txs = block['block']['data']['txs']
            txs_size = sum(base64_decoded_size(item) for item in block_txs)

            if previous_block_timestamp is not None:
                block_time = (block_timestamp - previous_block_timestamp)
                print(
                    f"{n - 1} to {n} :: {block_time} "
                    f"(at {block_timestamp}) (size {txs_size})"
                )

            previous_block_timestamp = block_timestamp
        except Exception as err:
            print(f"ERR : {err}")

        time.sleep(1)


# Report 1
for i in range(len(start_heights)):
    print(f"\n--- (Report 1.{i})")
    run_report_1(start_heights[i], end_heights[i])

# Report 2
for i, signer in enumerate(signers):
    print(f"\n--- (Report 2.{i})")
    for j in range(len(start_heights)):
        run_report_2(start_heights[j], end_heights[j], signer)

# Report 3
for i in range(len(start_heights)):
    print(f"\n--- (Report 3.{i})")
    run_report_3(start_heights[i], end_heights[i])
