import os
import base64
import json
import sys

from utils.classes import FuelSequencerChain
from utils.constants import *

import logging
import logging.handlers
import multiprocessing
import time


# Issue transactions
def process_block(queue, rollup, current_block, topic, order):
    logger = logging.getLogger(f"{rollup}")
    handler = logging.handlers.QueueHandler(queue)
    logger.addHandler(handler)
    logger.setLevel(logging.DEBUG)

    logger.info(f"processing block {current_block} for rollup {rollup} using topic {topic} order {order}")

    data = base64.b64encode(os.urandom(DATA_SIZE)).decode('utf-8')
    SEQ.post_blob(sender, topic, f"{order}", data, gas, fee, f"{os.path.dirname(os.path.abspath(__file__))}/txs")

    logger.info(f"finished processing block {current_block} for rollup {rollup} using topic {topic} order {order}")


SEQ_node = "https://rpc-seq.simplystaking.xyz"
SEQ_chain = "seq-devnet-3"
SEQ_bin = "/Users/dferendo-simply/go/bin/fuelsequencerd"

SEQ = FuelSequencerChain(
    binary=SEQ_bin,
    node=SEQ_node,
    chain_id=SEQ_chain,
    key_name=key_name_alice,
    voting_period=10,
    fee_token="utest",
    gov_voters=["alice"],
)
SEQ.wait_for_txs = False

# Post Blob
sender = SEQ.address_alice
# maybe dependent on the size
gas = "1000000"
fee = [{"amount": "25000", "denom": SEQ.fee_token}]

BLOCKS_TO_LOAD_TEST = 3
DATA_SIZE = 1024
NUMBER_OF_PROCESSES = 1
TOPICS = {}

for i in range(NUMBER_OF_PROCESSES):
    # Topic, Order
    TOPICS[i] = (base64.b64encode(os.urandom(32)).decode('utf-8'), 0)

if __name__ == "__main__":
    queue = multiprocessing.Queue()
    listener = logging.handlers.QueueListener(queue, logging.StreamHandler())
    listener.start()

    starting_block = SEQ.query_last_block_height()
    processed_blocks = set()

    processes = []

    while True:
        current_block = SEQ.query_last_block_height()

        if current_block not in processed_blocks:
            processed_blocks.add(current_block)

            # Issue a thread (mimicking a rollup)
            for i in range(NUMBER_OF_PROCESSES):
                # Processing
                process = multiprocessing.Process(target=process_block,
                                                  args=(queue, i, current_block, TOPICS[i][0], TOPICS[i][1]))
                process.start()
                processes.append(process)
                TOPICS[i] = (TOPICS[i][0], TOPICS[i][1] + 1)

            # Stop submitting txs
            if current_block >= starting_block + BLOCKS_TO_LOAD_TEST:
                break

        # wait for new block to be produced
        time.sleep(0.5)

    # Wait for all processes to finish
    for process in processes:
        process.join()

    print("Finished issuing txs")

    # Wait until transactions are included
    time.sleep(12)
    ending_block = SEQ.query_last_block_height()

    # Monitoring
    for height in range(starting_block, ending_block):
        block = json.loads(SEQ.query_block(height))
        txs = block.get('data', {}).get('txs', [])

        print(f"block {height} had size {sys.getsizeof(txs)} bytes")
