import time

from dateutil import parser
import requests

sequencer_rpc = "https://rpc-seq.simplystaking.xyz/"
block_query = "block?height="


def get_block(n):
    return requests.get(f"{sequencer_rpc}{block_query}{n}").json()['result']


def get_latest_block():
    return get_block("")


latest_block = get_latest_block()
latest_height = int(latest_block['block']['header']['height'])
print(f"Latest height: {latest_height}")
start_height = latest_height - 1000
print(f"Start height: {latest_height}")

# Optional: manual override of heights
# start_height =
# latest_height =

previous_block_timestamp = None
for n in range(start_height, latest_height):

    try:
        block = get_block(n)
        block_timestamp = parser.parse(block['block']['header']['time'])

        if previous_block_timestamp is not None:
            block_time = (block_timestamp - previous_block_timestamp)
            print(f"{n-1} to {n} :: {block_time} (at {block_timestamp})")

        previous_block_timestamp = block_timestamp
    except Exception as err:
        print(f"ERR : {err}")

    time.sleep(1)
