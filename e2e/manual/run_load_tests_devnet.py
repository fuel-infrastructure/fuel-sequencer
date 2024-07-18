import base64
import json
import logging
import logging.handlers
import multiprocessing
import os
import sys
import time
from typing import Tuple

from utils.classes import FuelSequencerChain
from utils.constants import *

EXPLORER_TX_URL = "https://seq.simplystaking.xyz/fuel/tx/"
SEQ_node = "https://rpc-seq.simplystaking.xyz"
SEQ_chain = "seq-devnet-4"
SEQ_bin = "fuelsequencerd"  # needs to be in $GOPATH/bin

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
SEQ.gas_prices = f"10000000000{SEQ.fee_token}"

# Load test configuration
MAX_TEMP_FILES = 10
BLOCKS_TO_LOAD_TEST = 1
BLOB_SIZE_BYTES = 700000
# Max BLOB_SIZE_BYTES: 1048576
# Ref: https://rest-seq.simplystaking.xyz/fuelsequencer/sequencing/v1/params
print(f"Running for {BLOCKS_TO_LOAD_TEST} blocks "
      f"with blobs of {BLOB_SIZE_BYTES} bytes")

# MsgPostBlob transaction configuration
sender = SEQ.address_alice
gas = 100000 + (10 * BLOB_SIZE_BYTES)  # 10 = tx_size_cost_per_byte
fee_amount = int(gas) * int(SEQ.gas_prices.replace(SEQ.fee_token, ""))
fee = [{"amount": f"{fee_amount}", "denom": SEQ.fee_token}]


def get_temp_txs_folder():
    return os.path.join(os.path.dirname(os.path.abspath(__file__)), "temp_txs")


def get_temp_tx_files(rollup, order) -> Tuple[str, str]:
    tx_file_path = os.path.join(
        get_temp_txs_folder(),
        f"temp.{rollup}.{order % MAX_TEMP_FILES}.json"
    )
    result_file_path = os.path.join(
        get_temp_txs_folder(),
        f"temp.{rollup}.{order % MAX_TEMP_FILES}-result.json"
    )
    return tx_file_path, result_file_path


def print_block_size(height: int):
    block = json.loads(SEQ.query_block(height))
    txs = block.get('data', {}).get('txs', [])

    block_size = sum(sys.getsizeof(item) for item in txs)
    print(f"Block {height} had size {block_size} bytes")


# Issue transactions
def post_blob(
        logging_queue: multiprocessing.Queue,
        rollup: int,
        block: int,
        topic: str,
        order: int,
        acc_num: int,
        acc_starting_seq: int,
):
    logger = logging.getLogger(f"{rollup}")
    handler = logging.handlers.QueueHandler(logging_queue)
    formatter = logging.Formatter(
        '%(asctime)s %(levelname)-6s %(message)s',
        datefmt='%Y-%m-%d %H:%M:%S',
    )
    handler.setFormatter(formatter)
    logger.addHandler(handler)
    logger.setLevel(logging.DEBUG)

    logger.info(
        f"processing block={block} rollup={rollup} topic={topic} order={order}"
    )

    data = base64.b64encode(os.urandom(BLOB_SIZE_BYTES)).decode('utf-8')
    tx_file_path, result_file_path = get_temp_tx_files(rollup, order)

    # Configure offline signing...
    # Sequence is the starting sequence plus the topic order
    # e.g. for start sequence 100, and order 5, the sequence will be 105
    SEQ.offline_signing = True
    SEQ.account_number = acc_num
    SEQ.account_sequence = int(acc_starting_seq) + int(order)

    result = SEQ.post_blob_from_file(
        sender, topic, f"{order}", data, gas, fee, tx_file_path
    )
    with open(result_file_path, 'w') as f:
        f.write(result)

    # Try to parse tx hash
    try:
        result = json.loads(result)
        if result['code'] != 0:
            tx_info = result['raw_log']
        else:
            tx_info = f"{EXPLORER_TX_URL}{result['txhash']}"
    except Exception:
        tx_info = "ERR"

    logger.info(
        f"finished block={block} rollup={rollup} "
        f"topic={topic} order={order} acc_num={SEQ.account_number} "
        f"acc_seq={SEQ.account_sequence} tx :: {tx_info}"
    )


if __name__ == "__main__":
    # Get account number and starting sequence
    account = json.loads(SEQ.query_account(sender))['account']['value']
    acc_num = account['account_number'] if 'account_number' in account else 0
    acc_starting_seq = account['sequence']

    # Pre-test cleanup
    for file in os.listdir(get_temp_txs_folder()):
        if file.startswith("temp"):
            os.remove(os.path.join(get_temp_txs_folder(), file))
    print("Cleaned up previous test files")

    # Generate a topic with a unique ID and starting order of 0
    topic_id = base64.b64encode(os.urandom(32)).decode('utf-8')
    topic_order = 0

    logging_queue = multiprocessing.Queue()
    listener = logging.handlers.QueueListener(
        logging_queue, logging.StreamHandler()
    )
    listener.start()

    # Calculate block range that test will run for
    start_block = SEQ.query_last_block_height()
    last_block = start_block + BLOCKS_TO_LOAD_TEST - 1
    prev_block = start_block - 1

    process = None
    while True:
        curr_block = SEQ.query_last_block_height()

        # Wait for a new block
        if curr_block == prev_block:
            time.sleep(0.5)
            continue

        print(f'Detected block={curr_block}')

        # Processing
        process = multiprocessing.Process(
            target=post_blob,
            args=(
                logging_queue, 0, curr_block,
                topic_id, topic_order, acc_num, acc_starting_seq,
            )
        )
        process.start()

        topic_order += 1

        # Last block reached
        if curr_block >= last_block:
            break

        prev_block = curr_block

    # Wait for last process to finish
    if process:
        process.join()

    print("Done")
