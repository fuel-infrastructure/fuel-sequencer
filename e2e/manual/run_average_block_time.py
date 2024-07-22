import time

import requests
from dateutil import parser

sequencer_rpc = "https://rpc-seq.simplystaking.xyz/"
block_query = "block?height="


def get_block(n):
    return requests.get(f"{sequencer_rpc}{block_query}{n}").json()['result']


def get_latest_block():
    return get_block("")


# Calculate block range based on latest height
latest_block = get_latest_block()
latest_height = int(latest_block['block']['header']['height'])
end_height = latest_height
start_height = end_height - 100

# Optional: manual override of heights
# start_height =
# latest_height =

print(f"Start height: {start_height}")
print(f"End height: {end_height}")


def base64_decoded_size(encoded_str):
    l = len(encoded_str)
    p = encoded_str.count('=')  # Count padding characters
    return (l * 3) // 4 - p


# Calculate overall average
start_block = get_block(start_height)
start_block_timestamp = parser.parse(start_block['block']['header']['time'])
end_block = get_block(end_height)
end_block_timestamp = parser.parse(end_block['block']['header']['time'])
block_range_time = (end_block_timestamp - start_block_timestamp)
seconds_per_block = block_range_time / (end_height - start_height)
print(
    f"From {start_height} to {end_height}:\n"
    f"\tNumber of blocks: {end_height - start_height}\n"
    f"\tTime elapsed: {block_range_time}\n"
    f"\tSeconds per block: {seconds_per_block}"
)

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
