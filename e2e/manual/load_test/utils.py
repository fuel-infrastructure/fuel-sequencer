import logging
import logging.handlers
import multiprocessing
import os
from typing import Tuple, Dict

MAX_TEMP_FILES = 10


def get_temp_txs_folder():
    return "temp_txs"


def get_temp_tx_file_prefix(key_name: str):
    return f"temp.{key_name}"


def get_temp_tx_files(key_name: str, rollup: int, order: int) -> Tuple[str, str]:
    filename_prefix = get_temp_tx_file_prefix(key_name)
    tx_file_path = os.path.join(
        get_temp_txs_folder(),
        f"{filename_prefix}.{rollup}.{order % MAX_TEMP_FILES}.json",
    )
    result_file_path = os.path.join(
        get_temp_txs_folder(),
        f"{filename_prefix}.{rollup}.{order % MAX_TEMP_FILES}-result.json",
    )
    return tx_file_path, result_file_path


def get_logger(name: str, logging_queue: multiprocessing.Queue) -> logging.Logger:
    logger = logging.getLogger(name)
    handler = logging.handlers.QueueHandler(logging_queue)
    formatter = logging.Formatter(
        "%(asctime)s %(levelname)-6s %(message)s",
        datefmt="%Y-%m-%d %H:%M:%S",
    )
    handler.setFormatter(formatter)
    logger.addHandler(handler)
    logger.setLevel(logging.DEBUG)
    return logger


def get_tx_info_from_tx_result(result: Dict, explorer_tx_url: str) -> str:
    try:
        if result["code"] != 0:
            return result["raw_log"]
        else:
            return f"{explorer_tx_url}{result['txhash']}"
    except Exception:
        return "ERR"
