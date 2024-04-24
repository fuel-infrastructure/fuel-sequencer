import base64
import json
import subprocess
import time
from typing import List, Optional, Dict, Union, Type

import requests
from utils.constants import events_filter, events_filter_by_prefix
from web3 import Web3, HTTPProvider
from web3.contract import Contract

from raw_msgs.msg_post_blob import get_msg_post_blob


class CosmosChain:
    def __init__(
            self,
            binary: str,
            node: str,
            chain_id: str,
            key_name: str,
            voting_period: int,
            fee_token: str,
            gov_voters: List[str],
    ):
        self.binary = binary
        self.node = node
        self.node_http = node.replace("tcp://", "http://")
        self.chain_id = chain_id
        self.key_name = key_name
        self.voting_period = voting_period
        self.voting_period_padding = 1
        self.fee_token = fee_token

        self.output = "json"
        self.keyring_backend = "test"
        self.gas = "auto"
        self.gas_adjustment = "1.5"
        self.gas_prices = f"0.025{fee_token}"
        self.broadcast_mode = "sync"
        self.wait_for_txs = True

        # Older chains use base64 events
        self.base64_events = True

        # Voters (list of key names) needed to vote YES for a proposal to pass
        self.gov_voters = gov_voters

        self.errors_enabled = True

    def _run_command(self, command: str):
        stderr = None if self.errors_enabled else subprocess.DEVNULL
        process = subprocess.Popen(command.split(),
                                   stdout=subprocess.PIPE,
                                   stderr=stderr)
        output, error = process.communicate()
        output = output.decode('utf-8')
        if output.endswith("\n"):
            output = output[:-1]
        return output

    def _run_piped_commands(self, command1: str, command2: str):
        stderr = None if self.errors_enabled else subprocess.DEVNULL
        process1 = subprocess.Popen(command1.split(),
                                    stdout=subprocess.PIPE,
                                    stderr=stderr)
        output = subprocess.check_output(command2.split(),
                                         stdin=process1.stdout,
                                         stderr=stderr)
        process1.wait()
        output = output.decode('utf-8')
        if output.endswith("\n"):
            output = output[:-1]
        return output

    def tx(self, command: str, signer_key: str = "", wait_for_txs=True):
        if signer_key == "":
            signer_key = self.key_name
        tx = self._run_command(
            f"{self.binary} tx {command} "
            f"--from={signer_key} "
            f"--chain-id={self.chain_id} "
            f"--node={self.node} "
            f"--keyring-backend={self.keyring_backend} "
            f"-b={self.broadcast_mode} "
            f"-o={self.output} "
            f"--gas={self.gas} "
            f"--gas-adjustment={self.gas_adjustment} "
            f"--gas-prices={self.gas_prices} "
            "-y")
        if (self.broadcast_mode == 'block' or
                not self.wait_for_txs or
                not wait_for_txs):
            return tx
        else:
            tx_hash = json.loads(tx)["txhash"]
            return self.wait_for_tx(tx_hash)

    def sign(self, file: str):
        return self.tx(f"sign {file} --output-document={file}", wait_for_txs=False)

    def broadcast(self, file: str):
        return self.tx(f"broadcast {file}")

    def keys(self, command: str):
        return self._run_command(
            f"{self.binary} keys {command} "
            f"--keyring-backend={self.keyring_backend}")

    def query(self, command: str):
        return self._run_command(
            f"{self.binary} q {command} "
            f"--node={self.node} "
            f"-o={self.output}")

    def get_json(self, endpoint: str):
        return requests.get(f"{self.node_http}{endpoint}").json()

    def add_key(self, key_name: str, mnemonic: str, with_delete=True) -> str:
        print(f"Adding key: {key_name}")

        if with_delete:
            errors_enabled = self.errors_enabled
            self.errors_enabled = False
            self.keys(f"delete {key_name} "
                      "--keyring-backend={self.keyring_backend} "
                      "-y")
            self.errors_enabled = errors_enabled

        command1 = f"echo {mnemonic}"
        command2 = f"{self.binary} keys add {key_name} " \
                   f"--recover " \
                   f"--keyring-backend={self.keyring_backend}"
        return self._run_piped_commands(command1, command2)

    def add_key_manually(self, name: str, mnemonic: str,
                         with_delete=False) -> None:
        delete = (f'{self.binary} keys delete {name} -y '
                  f'--keyring-backend=test '
                  f'&& ') if with_delete else ''
        add_key = f'{delete}' \
                  f'echo "{mnemonic}" | ' \
                  f'{self.binary} keys add {name} --recover ' \
                  '--keyring-backend=test '
        print(add_key)

    def add_keys(
            self, names: List[str], mnemonics: List[str], with_delete=True
    ) -> None:
        for i in range(len(names)):
            self.add_key(names[i], mnemonics[i], with_delete)

    def add_keys_manually(
            self, names: List[str], mnemonics: List[str], with_delete=False
    ) -> None:
        for i in range(len(names)):
            self.add_key_manually(names[i], mnemonics[i], with_delete)

    def submit_gov_proposal(
            self,
            proposal_dict: Dict,
            vote: bool = True
    ) -> None:
        temp_json_file = "temp-proposal.json"
        with open(temp_json_file, 'w') as f:
            json.dump(proposal_dict, f)

        # Submit proposal
        output = self.tx(f"gov submit-proposal {temp_json_file}")
        print(f"SUBMITTED PROPOSAL {output}")

        # Get proposal ID
        json_output = json.loads(output)
        events = json_output["events"]
        submit_prop = [e for e in events if e["type"] == "submit_proposal"][0]
        proposal_id = \
        [a for a in submit_prop["attributes"] if a["key"] == "proposal_id"][0][
            "value"]
        print(f"PROPOSAL ID: {proposal_id}")

        if not vote:
            return proposal_id

        # Vote yes
        for voter in self.gov_voters:
            output = self.tx(f"gov vote {proposal_id} yes", voter)
            print(f"VOTED YES ({voter}): {output}")

        # Sleep
        sleep = self.voting_period + self.voting_period_padding
        print(f"SLEEPING FOR {sleep} SECONDS...")
        time.sleep(sleep)

        # Query proposal
        output = self.query(f"gov proposal {proposal_id}")
        print(f"PROPOSAL: {output}")
        json_output = json.loads(output)
        print(f"PROPOSAL STATUS: {json_output['status']}")

        return proposal_id

    def submit_update_client_proposal_legacy(
            self,
            subject_client_id: str,
            substitute_client_id: str,
            title: str = "Title",
            description: str = "Description",
            vote: bool = True,
    ) -> None:

        # Submit proposal
        output = self.tx(
            f"gov submit-legacy-proposal update-client "
            f"{subject_client_id} {substitute_client_id} "
            f"--title={title} --description={description} "
            f"--deposit=10000000{self.fee_token}")
        print(f"SUBMITTED PROPOSAL {output}")

        # Get proposal ID
        json_output = json.loads(output)
        raw_log = json_output["raw_log"]
        proposal_id_index = raw_log.index("proposal_id")
        proposal_id_start = proposal_id_index + 22
        proposal_id_end_offset = raw_log[proposal_id_start:].index('"')
        proposal_id_end = proposal_id_start + proposal_id_end_offset
        proposal_id = raw_log[proposal_id_start:proposal_id_end]
        if proposal_id.endswith('\"'):
            proposal_id = proposal_id[:-1]
        print(f"PROPOSAL ID: {proposal_id}")

        if not vote:
            return proposal_id

        # Vote yes
        for voter in self.gov_voters:
            output = self.tx(f"gov vote {proposal_id} yes", voter)
            print(f"VOTED YES ({voter}): {output}")

        # Sleep
        sleep = self.voting_period + self.voting_period_padding
        print(f"SLEEPING FOR {sleep} SECONDS...")
        time.sleep(sleep)

        # Query proposal
        output = self.query(f"gov proposal {proposal_id}")
        print(f"PROPOSAL: {output}")
        json_output = json.loads(output)
        print(f"PROPOSAL STATUS: {json_output['status']}")

        return proposal_id

    def submit_param_change_proposal_legacy(
            self,
            proposal_dict: Dict,
            vote: bool = True,
    ) -> None:
        return self.submit_gov_proposal_legacy(
            "param-change", proposal_dict, vote
        )

    def submit_gov_proposal_legacy(
            self,
            proposal_type: str,
            proposal_dict: Dict,
            vote: bool = True,
    ) -> None:
        temp_json_file = "temp-proposal.json"
        with open(temp_json_file, 'w') as f:
            json.dump(proposal_dict, f)

        # Submit proposal
        output = self.tx(
            f"gov submit-legacy-proposal {proposal_type} {temp_json_file}")
        print(f"SUBMITTED PROPOSAL {output}")

        # Get proposal ID
        json_output = json.loads(output)
        raw_log = json_output["raw_log"]
        proposal_id_index = raw_log.index("proposal_id")
        proposal_id_start = proposal_id_index + 22
        proposal_id_end_offset = raw_log[proposal_id_start:].index('"')
        proposal_id_end = proposal_id_start + proposal_id_end_offset
        proposal_id = raw_log[proposal_id_start:proposal_id_end]
        if proposal_id.endswith('\"'):
            proposal_id = proposal_id[:-1]
        print(f"PROPOSAL ID: {proposal_id}")

        if not vote:
            return proposal_id

        # Vote yes
        for voter in self.gov_voters:
            output = self.tx(f"gov vote {proposal_id} yes", voter)
            print(f"VOTED YES ({voter}): {output}")

        # Sleep
        sleep = self.voting_period + self.voting_period_padding
        print(f"SLEEPING FOR {sleep} SECONDS...")
        time.sleep(sleep)

        # Query proposal
        output = self.query(f"gov proposal {proposal_id}")
        print(f"PROPOSAL: {output}")
        json_output = json.loads(output)
        print(f"PROPOSAL STATUS: {json_output['status']}")

        return proposal_id

    def submit_param_change_proposal(
            self,
            proposal_dict: Dict,
            vote: bool = True
    ) -> None:
        temp_json_file = "temp-proposal.json"
        with open(temp_json_file, 'w') as f:
            json.dump(proposal_dict, f)

        # Submit proposal
        output = self.tx(
            f"gov submit-proposal param-change {temp_json_file}")
        print(f"SUBMITTED PROPOSAL {output}")

        # Get proposal ID
        json_output = json.loads(output)
        raw_log = json_output["raw_log"]
        proposal_id_index = raw_log.index("proposal_id")
        proposal_id_start = proposal_id_index + 22
        proposal_id_end_offset = raw_log[proposal_id_start:].index('"')
        proposal_id_end = proposal_id_start + proposal_id_end_offset
        proposal_id = raw_log[proposal_id_start:proposal_id_end]
        if proposal_id.endswith('\"'):
            proposal_id = proposal_id[:-1]
        print(f"PROPOSAL ID: {proposal_id}")

        if not vote:
            return proposal_id

        # Vote yes
        for voter in self.gov_voters:
            output = self.tx(f"gov vote {proposal_id} yes", voter)
            print(f"VOTED YES ({voter}): {output}")

        # Sleep
        sleep = self.voting_period + self.voting_period_padding
        print(f"SLEEPING FOR {sleep} SECONDS...")
        time.sleep(sleep)

        # Query proposal
        output = self.query(f"gov proposal {proposal_id}")
        print(f"PROPOSAL: {output}")
        json_output = json.loads(output)
        print(f"PROPOSAL STATUS: {json_output['status']}")

        return proposal_id

    def submit_group_proposal(self, proposal_dict: Dict) -> None:
        temp_json_file = "temp-proposal.json"
        with open(temp_json_file, 'w') as f:
            json.dump(proposal_dict, f)

        # Submit proposal
        output = self.tx(
            f"group submit-proposal --exec=try {temp_json_file}")
        print(f"SUBMITTED PROPOSAL {output}")

        # Get proposal ID
        json_output = json.loads(output)
        raw_log = json_output["raw_log"]
        proposal_id_index = raw_log.index("proposal_id")
        proposal_id_start = proposal_id_index + 24
        proposal_id_end_offset = raw_log[proposal_id_start:].index('"')
        proposal_id_end = proposal_id_start + proposal_id_end_offset
        proposal_id = raw_log[proposal_id_start:proposal_id_end]
        if proposal_id.endswith('\\'):
            proposal_id = proposal_id[:-1]
        print(f"PROPOSAL ID: {proposal_id}")

        if 'PROPOSAL_EXECUTOR_RESULT_SUCCESS' in output:
            print(f"PROPOSAL STATUS: PROPOSAL_STATUS_ACCEPTED")
            print(f"PROPOSAL EXECUTION: PROPOSAL_EXECUTOR_RESULT_SUCCESS")
            return proposal_id

        # Query proposal
        output = self.query(f"group proposal {proposal_id}")
        print(f"PROPOSAL: {output}")
        json_output = json.loads(output)
        print(f"PROPOSAL STATUS: {json_output['proposal']['status']}")
        executor_result = json_output['proposal']['executor_result']
        print(f"PROPOSAL EXECUTION: {executor_result}")

        return proposal_id

    def send(self, receiver_addr: str, amount: str) -> str:
        return self.tx(
            f"bank send {self.key_name} {receiver_addr} {amount}")

    def ibc_transfer(self, channel_id: str, receiver_addr: str,
                     amount: str) -> str:
        return self.tx(
            f"ibc-transfer transfer transfer {channel_id} {receiver_addr} {amount}")

    def query_tx(self, tx_hash: str):
        return self.query(f"tx {tx_hash}")

    def query_block(self, height: int):
        return self.query(f"block {height} --type=height")

    def query_balance_by_key_name(self, key_name: str) -> str:
        output = self.keys(f"show {key_name} -a")
        return self.query(f"bank balances {output}")

    def query_balance_by_address(self, address: str) -> str:
        return self.query(f"bank balances {address}")

    def query_denom_trace(self, hash_or_denom: str) -> str:
        return self.query(f"ibc-transfer denom-trace {hash_or_denom}")

    def query_module_params(self, module: str) -> str:
        return self.query(f"{module} params")

    @staticmethod
    def _decode_events(events: Optional[List], base64_events: bool) -> List:
        if events is None:
            return []

        new_events = []
        for event in events:
            if event['type'] in events_filter:
                continue
            skip = False
            for prefix in events_filter_by_prefix:
                if event['type'].startswith(prefix):
                    skip = True
                    break
            if skip:
                continue

            attr = event['attributes']
            for i in range(len(attr)):

                if base64_events:
                    new_key = base64.standard_b64decode(attr[i]['key']).decode(
                        'utf-8')
                else:
                    new_key = attr[i]['key']
                if attr[i]['value'] is None:
                    new_value = 'None'
                else:
                    if base64_events:
                        new_value = base64.standard_b64decode(
                            attr[i]['value']).decode('utf-8')
                    else:
                        new_value = attr[i]['value']

                attr[i]['key'] = new_key
                attr[i]['value'] = new_value

            new_events.append({'type': event['type'], 'attributes': attr})

        return new_events

    def get_block(self, height: int):
        while True:
            output = self.get_json(f'/block?height={height}')
            if 'error' in output and 'height' in output['error']['data']:
                time.sleep(1)
            else:
                break

        return output['result']

    def get_last_block_height(self) -> int:
        abci_info = self.get_json('/abci_info')
        return int(abci_info['result']['response']['last_block_height'])

    def get_block_events(self, height: int):
        if height <= 0:
            raise Exception("height must be > 0")

        while True:
            output = self.get_json(f'/block_results?height={height}')
            if 'error' in output and 'height' in output['error']['data']:
                time.sleep(1)
            else:
                break

        result = output['result']

        tx_events = []
        if result['txs_results'] is not None:
            for tx_result in result['txs_results']:
                msgs_events = tx_result['events']
                for msg_events in msgs_events:
                    tx_events.append(msg_events)

        return {
            'BB': self._decode_events(result['begin_block_events'],
                                      self.base64_events),
            'EB': self._decode_events(result['end_block_events'],
                                      self.base64_events),
            'TX': self._decode_events(tx_events,
                                      self.base64_events),
        }

    def query_channel_end(self, port: str, channel_id: str):
        return self.query(f"ibc channel end {port} {channel_id}")

    def query_channels(self):
        return self.query(f"ibc channel channels")

    def query_packet_commitments(self, port: str, channel_id: str):
        return self.query(f"ibc channel packet-commitments {port} {channel_id}")

    def query_unreceived_acks(self, port: str, channel_id: str):
        return self.query(f"ibc channel unreceived-acks {port} {channel_id}")

    def query_unreceived_packets(self, port: str, channel_id: str):
        return self.query(f"ibc channel unreceived-packets {port} {channel_id}")

    def query_connection_end(self, connection_id: str):
        return self.query(f"ibc connection end {connection_id}")

    def query_client_status(self, client_id: str):
        return self.query(f"ibc client status {client_id}")

    def query_proposal(self, proposal_id: str):
        return self.query(f"gov proposal {proposal_id}")

    def query_account(self, address: str):
        return self.query(f"auth account {address}")

    def query_grants_by_grantee(self, grantee_addr: str):
        return self.query(f"authz grants-by-grantee {grantee_addr}")

    def query_grants_by_granter(self, granter_addr: str):
        return self.query(f"authz grants-by-granter {granter_addr}")

    def base_denom_from_ibc_denom(self, ibc_denom: str):
        return json.loads(
            self.query_denom_trace(ibc_denom))['denom_trace']['base_denom']

    def query_community_pool(self):
        return self.query("distribution community-pool")

    def wait_for_tx(self, tx_hash: str) -> str:
        # Stash errors_enabled and set actual one to False
        errors_enabled = self.errors_enabled
        self.errors_enabled = False

        while True:
            try:
                tx = self.query_tx(tx_hash)
                _ = json.loads(tx)
                break
            except Exception as e:
                time.sleep(0.5)

        # restore errors_enabled
        self.errors_enabled = errors_enabled

        return tx


class FuelSequencerChain(CosmosChain):
    governance_address = "fuelsequencer10d07y265gmmuvt4z0w9aw880jnsr700jdjfvk3"

    address_alice = "fuelsequencer1vtfzrk6f4m6kxt6ehyqt9j5su5hvcz5q3dmlsm"
    address_bob = "fuelsequencer163rsv65t4893t2rz5rmda9sly7lgdlq2jgr36m"
    address_charlie = "fuelsequencer1n79wsstpakv0gw2efmruf9x8xs9m4rqfazfu8g"
    address_dexter = "fuelsequencer1r8aaf8fjcft7h7tupnafyv0kzk0342gnjls2pu"
    address_eve = "fuelsequencer13hfdkxj5aeqzsll569mqreedkafp34ngcsjkqjmpq6prtgv80kcq83gttw"

    def query_bridge_params(self) -> str:
        return self.query_module_params("bridge")

    def query_sequencing_params(self) -> str:
        return self.query_module_params("sequencing")

    def query_last_ethereum_block_synced(self) -> str:
        return self.query("bridge show-last-ethereum-block-synced")

    def query_ethereum_event_index_offset(self) -> str:
        return self.query("bridge show-ethereum-event-index-offset")

    def query_seq_address_from_eth_address(self, seq_address: str) -> str:
        return json.loads(self.query(
            f"bridge sequencer-address-from-ethereum-address {seq_address}"
        ))['sequencer_address']

    def query_topics(self) -> str:
        return self.query("sequencing list-topic")

    def query_topic(self, topic_id: str) -> str:
        return self.query(f"sequencing show-topic {topic_id}")

    def withdraw(self, to: str, amount: str) -> str:
        return self.tx(f"bridge withdraw-to-ethereum {to} {amount}")

    def post_blob(
            self,
            sender: str,
            topic: str,
            order: str,
            data: str,
            gas: str,
            fee: List
    ):
        msg = get_msg_post_blob(sender, topic, order, data, gas, fee)
        temp_json_file = "temp-msg.json"
        with open(temp_json_file, 'w') as f:
            json.dump(msg, f)

        self.sign(temp_json_file)
        return self.broadcast(temp_json_file)



class EthereumChain(Web3):

    def __init__(
            self,
            httpProvider: HTTPProvider,
            fuelstreamx_address: str,
            fuelstreamx_abi: str,
            acc_private_key: str,
            acc_address: str,
    ):
        super().__init__(httpProvider)

        self.fuelstreamx_address = fuelstreamx_address
        self.fuelstreamx_abi = fuelstreamx_abi
        self.acc_address = acc_address
        self.acc_private_key = acc_private_key

    # noinspection PyTypeChecker
    def _contract(self) -> Union[Type[Contract], Contract]:
        return self.eth.contract(
            address=self.fuelstreamx_address,
            abi=self.fuelstreamx_abi,
        )

    def _sign_tx(self, txn):
        return self.eth.account.sign_transaction(
            txn, private_key=self.acc_private_key,
        )

    # noinspection PyTypeChecker
    def deposit(self, amount: int, to: str, duration: int):
        # NB: function name is case-sensitive.
        txn = self._contract().functions.deposit(
            amount, to, duration
        ).build_transaction({
            'nonce': self.eth.get_transaction_count(self.acc_address),
        })

        signed_txn = self._sign_tx(txn)
        return self.eth.send_raw_transaction(signed_txn.rawTransaction)

    # noinspection PyTypeChecker
    def authorize(self, hex_bytes: str):
        # NB: function name is case-sensitive.
        txn = self._contract().functions.Authorize(
            hex_bytes,
        ).build_transaction({
            'nonce': self.eth.get_transaction_count(self.acc_address),
        })

        signed_txn = self._sign_tx(txn)
        return self.eth.send_raw_transaction(signed_txn.rawTransaction)
