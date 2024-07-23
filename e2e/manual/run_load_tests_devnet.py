import argparse
import base64
import json
import logging.handlers
import multiprocessing
import sys
import time

from load_test.post_blob import PostBlobRequest, post_blob_async
from load_test.utils import *
from utils.classes import FuelSequencerChain
from utils.constants import *

EXPLORER_TX_URL = "https://seq.simplystaking.xyz/fuel/tx/"
SEQ_node = "https://rpc-seq.simplystaking.xyz"
SEQ_chain = "seq-devnet-4"
SEQ_bin = "fuelsequencerd"  # needs to be in $GOPATH/bin

# Load test configuration
BLOCKS_TO_LOAD_TEST = 999999
SHUT_DOWN_ON_TX_ERR = True
BLOB_SIZE_BYTES = 1000000
# Max blob size:
#   https://rest-seq.simplystaking.xyz/fuelsequencer/sequencing/v1/params
# Max block size:
#   https://rpc-seq.simplystaking.xyz/consensus_params

if __name__ == "__main__":
    # Create the parser
    parser = argparse.ArgumentParser(description="Parse key and mnemonic")

    # Add arguments
    parser.add_argument('--key', type=str, required=False, help='key name')
    parser.add_argument('--mnemonic', type=str, required=False, help='mnemonic')

    # Parse the arguments
    args = parser.parse_args()
    if args.key and args.mnemonic:
        print("Using supplied key and mnemonic")
        key = args.key
        mnemonic = args.mnemonic
    elif not (args.key or args.mnemonic):
        print("Defaulting to alice's key and mnemonic")
        key = key_name_alice
        mnemonic = mnemonic_alice
    else:
        sys.exit("must set both or none of --key and --mnemonic")

    seq = FuelSequencerChain(
        binary=SEQ_bin,
        node=SEQ_node,
        chain_id=SEQ_chain,
        key_name=key,
        voting_period=10,
        fee_token="utest",
        gov_voters=["<unused>"],
    )
    seq.wait_for_txs = False
    seq.gas_prices = f"10000000000{seq.fee_token}"

    # MsgPostBlob transaction configuration
    gas = 100000 + (10 * BLOB_SIZE_BYTES)  # 10 = tx_size_cost_per_byte
    fee_amount = int(int(gas) * float(seq.gas_prices.replace(seq.fee_token, "")))
    fee = [{"amount": f"{fee_amount}", "denom": seq.fee_token}]

    # Ensure key is in place
    seq.add_keys(names=[key], mnemonics=[mnemonic])

    stop_event = multiprocessing.Event()

    print(f"Running load test for {BLOCKS_TO_LOAD_TEST} blocks "
          f"with blobs of {BLOB_SIZE_BYTES} bytes")

    # Get account number and starting sequence
    sender = seq.keys(f"show {key} -a")
    account = json.loads(seq.query_account(sender))['account']['value']
    acc_num = account['account_number'] if 'account_number' in account else 0
    acc_starting_seq = account['sequence'] if 'sequence' in account else 0

    # Pre-test cleanup
    filename_prefix = get_temp_tx_file_prefix(seq.key_name)
    for file in os.listdir(get_temp_txs_folder()):
        if file.startswith(filename_prefix):
            os.remove(os.path.join(get_temp_txs_folder(), file))
    print(f"Cleaned up previous test files ({filename_prefix}*)")

    # Generate a topic with a unique ID and starting order of 0
    topic_id = base64.b64encode(os.urandom(32)).decode('utf-8')
    topic_order = 0

    logging_queue = multiprocessing.Queue()
    listener = logging.handlers.QueueListener(
        logging_queue, logging.StreamHandler()
    )
    listener.start()

    # Calculate block range that test will run for
    start_block = seq.query_last_block_height()
    last_block = start_block + BLOCKS_TO_LOAD_TEST - 1
    prev_block = start_block - 1

    process = None
    while not (stop_event.is_set() and SHUT_DOWN_ON_TX_ERR):
        curr_block = seq.query_last_block_height()

        # Wait for a new block
        if curr_block == prev_block:
            time.sleep(0.5)
            continue

        print(f'Detected block={curr_block}')

        # Processing
        process = post_blob_async(
            PostBlobRequest(
                0,
                curr_block,
                topic_id,
                topic_order,
                acc_num,
                acc_starting_seq,
                BLOB_SIZE_BYTES,
                gas,
                fee,
                seq,
                EXPLORER_TX_URL,
            ),
            logging_queue,
            stop_event,
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
