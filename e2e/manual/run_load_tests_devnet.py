import json
import os
import base64

from utils.classes import FuelSequencerChain
from utils.constants import *

import multiprocessing
import time


# Issue transactions
def process_block(current_block, order):
    print("processing block", current_block)

    data = base64.b64encode(os.urandom(DATA_SIZE)).decode('utf-8')
    SEQ.post_blob(sender, topic_name, f"{order}", data, gas, fee)

    print("finished submitting for block ", current_block)


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

BLOCKS_TO_LOAD_TEST = 5
DATA_SIZE = 1024
# NUMBER_OF_PROCESSES = 4

topic_name = base64.b64encode(os.urandom(32)).decode('utf-8')

if __name__ == "__main__":
    starting_block = SEQ.query_last_block_height()
    order = 0
    processed_blocks = set()

    processes = []

    while True:
        current_block = SEQ.query_last_block_height()

        if current_block not in processed_blocks:
            processed_blocks.add(current_block)

            # Processing
            process = multiprocessing.Process(target=process_block, args=(current_block, order))
            process.start()
            processes.append(process)
            order += 1

            # Stop submitting txes
            if current_block >= starting_block + BLOCKS_TO_LOAD_TEST:
                break

        # wait for new block to be produced
        time.sleep(0.5)

    # Wait for all processes to finish
    for process in processes:
        process.join()

    print("Finished")
