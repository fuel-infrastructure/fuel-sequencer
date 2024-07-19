import base64
import json
import logging.handlers
import multiprocessing
import os
from typing import List

from load_test.utils import get_temp_tx_files


class PostBlobRequest:

    def __init__(
            self,
            rollup: int,
            block: int,
            topic_id: str,
            topic_order: int,
            acc_num: int,
            acc_starting_seq: int,
            blob_size_bytes: int,
            tx_gas: int,
            tx_fee: List,
            seq,  # FuelSequencerChain,
            explorer_tx_url: str,
    ):
        self.rollup = rollup
        self.block = block
        self.topic_id = topic_id
        self.topic_order = topic_order
        self.acc_num = acc_num
        self.acc_starting_seq = acc_starting_seq
        self.blob_size_bytes = blob_size_bytes
        self.tx_gas = tx_gas
        self.tx_fee = tx_fee
        self.seq = seq
        self.explorer_tx_url = explorer_tx_url


def post_blob(
        req: PostBlobRequest,
        logging_queue: multiprocessing.Queue,
        stop_event: multiprocessing.Event,
):
    logger = logging.getLogger(f"{req.rollup}")
    handler = logging.handlers.QueueHandler(logging_queue)
    formatter = logging.Formatter(
        '%(asctime)s %(levelname)-6s %(message)s',
        datefmt='%Y-%m-%d %H:%M:%S',
    )
    handler.setFormatter(formatter)
    logger.addHandler(handler)
    logger.setLevel(logging.DEBUG)

    logger.info(
        f"processing block={req.block} rollup={req.rollup} topic={req.topic_id} order={req.topic_order}"
    )

    data = base64.b64encode(os.urandom(req.blob_size_bytes)).decode('utf-8')
    tx_file_path, result_file_path = get_temp_tx_files(
        req.seq.key_name, req.rollup, req.topic_order
    )

    # Configure offline signing...
    # Sequence is the starting sequence plus the topic order
    # e.g. for start sequence 100, and order 5, the sequence will be 105
    req.seq.offline_signing = True
    req.seq.account_number = req.acc_num
    req.seq.account_sequence = int(req.acc_starting_seq) + int(req.topic_order)

    sender = req.seq.keys(f"show {req.seq.key_name} -a")
    result = req.seq.post_blob_from_file(
        sender, req.topic_id, f"{req.topic_order}", data, req.tx_gas,
        req.tx_fee, tx_file_path
    )
    with open(result_file_path, 'w') as f:
        f.write(result)

    # Try to parse tx hash
    try:
        result = json.loads(result)
        if result['code'] != 0:
            tx_info = result['raw_log']
        else:
            tx_info = f"{req.explorer_tx_url}{result['txhash']}"
    except Exception:
        tx_info = "ERR"

    logger.info(
        f"finished block={req.block} rollup={req.rollup} "
        f"topic={req.topic_id} order={req.topic_order} acc_num={req.seq.account_number} "
        f"acc_seq={req.seq.account_sequence} tx :: {tx_info}"
    )

    if result['code'] != 0:
        stop_event.set()


def post_blob_async(
        req: PostBlobRequest,
        logging_queue: multiprocessing.Queue,
        stop_event: multiprocessing.Event,
) -> multiprocessing.Process:
    return multiprocessing.Process(
        target=post_blob,
        args=(req, logging_queue, stop_event)
    )
