import os
from typing import Tuple

MAX_TEMP_FILES = 10


def get_temp_txs_folder():
    return "temp_txs"


def get_temp_tx_file_prefix(key_name: str):
    return f"temp.{key_name}"


def get_temp_tx_files(
        key_name: str,
        rollup: int,
        order: int
) -> Tuple[str, str]:
    filename_prefix = get_temp_tx_file_prefix(key_name)
    tx_file_path = os.path.join(
        get_temp_txs_folder(),
        f"{filename_prefix}.{rollup}.{order % MAX_TEMP_FILES}.json"
    )
    result_file_path = os.path.join(
        get_temp_txs_folder(),
        f"{filename_prefix}.{rollup}.{order % MAX_TEMP_FILES}-result.json"
    )
    return tx_file_path, result_file_path
