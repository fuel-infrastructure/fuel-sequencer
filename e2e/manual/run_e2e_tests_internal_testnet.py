import requests
from web3 import Web3

from proposals_gov.set_bridge_module_params import \
    get_update_bridge_module_params_proposal
from proposals_gov.community_pool_spend import \
    get_community_pool_spend_proposal
from proposals_gov.grant_authorisation import \
    get_grant_authorisation_proposal as get_grant_authorisation_proposal_by_gov
from proposals_gov.set_voting_period_low import \
    get_set_voting_period_low_proposal
from proposals_gov.software_upgrade import get_software_upgrade_proposal
from utils.classes import FuelSequencerChain, EthereumChain
from utils.constants import *
from utils.helpers import *


# Helper function to print JSON in a pretty way
def pretty(in_json: str):
    print(json.dumps(json.loads(in_json), indent=2))


SEQ_host = "80.64.208.225"
SEQ_port = "26657"
SEQ_chain = "fuelsequencer-test-2"
SEQ_bin = "fuelsequencerd"

ETH_rpc = "https://ethereum-holesky-rpc.publicnode.com"
ETH_fuelstreamx_address = "0x85d92cC1dB74b76E6C3d9F7CEf6c44967fe62A46"
ETH_fuelstreamx_abi = '[{"type":"constructor","inputs":[{"name":"_params","type":"tuple","internalType":"structFuelStreamX.InitParameters","components":[{"name":"guardian","type":"address","internalType":"address"},{"name":"gateway","type":"address","internalType":"address"},{"name":"height","type":"uint64","internalType":"uint64"},{"name":"header","type":"bytes32","internalType":"bytes32"},{"name":"nextHeaderFunctionId","type":"bytes32","internalType":"bytes32"},{"name":"headerRangeFunctionId","type":"bytes32","internalType":"bytes32"}]}],"stateMutability":"nonpayable"},{"type":"function","name":"Authorize","inputs":[{"name":"_message","type":"bytes","internalType":"bytes"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"BRIDGE_COMMITMENT_MAX","inputs":[],"outputs":[{"name":"","type":"uint64","internalType":"uint64"}],"stateMutability":"view"},{"type":"function","name":"VERSION","inputs":[],"outputs":[{"name":"","type":"string","internalType":"string"}],"stateMutability":"pure"},{"type":"function","name":"blockHeightToHeaderHash","inputs":[{"name":"","type":"uint64","internalType":"uint64"}],"outputs":[{"name":"","type":"bytes32","internalType":"bytes32"}],"stateMutability":"view"},{"type":"function","name":"commitHeaderRange","inputs":[{"name":"_targetBlock","type":"uint64","internalType":"uint64"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"commitNextHeader","inputs":[{"name":"_trustedBlock","type":"uint64","internalType":"uint64"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"deposit","inputs":[{"name":"_amount","type":"uint256","internalType":"uint256"},{"name":"_to","type":"string","internalType":"string"},{"name":"_duration","type":"uint256","internalType":"uint256"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"frozen","inputs":[],"outputs":[{"name":"","type":"bool","internalType":"bool"}],"stateMutability":"view"},{"type":"function","name":"gateway","inputs":[],"outputs":[{"name":"","type":"address","internalType":"address"}],"stateMutability":"view"},{"type":"function","name":"headerRangeFunctionId","inputs":[],"outputs":[{"name":"","type":"bytes32","internalType":"bytes32"}],"stateMutability":"view"},{"type":"function","name":"latestBlock","inputs":[],"outputs":[{"name":"","type":"uint64","internalType":"uint64"}],"stateMutability":"view"},{"type":"function","name":"nextHeaderFunctionId","inputs":[],"outputs":[{"name":"","type":"bytes32","internalType":"bytes32"}],"stateMutability":"view"},{"type":"function","name":"processSequencerWithdrawalMessage","inputs":[{"name":"_proofNonce","type":"uint256","internalType":"uint256"},{"name":"bridgeCommitmentLeaf","type":"tuple","internalType":"structFuelStreamX.BridgeCommitmentLeaf","components":[{"name":"height","type":"uint256","internalType":"uint256"},{"name":"resultsHash","type":"bytes32","internalType":"bytes32"}]},{"name":"bridgeCommitmentLeafProof","type":"tuple","internalType":"structBinaryMerkleProof","components":[{"name":"sideNodes","type":"bytes32[]","internalType":"bytes32[]"},{"name":"key","type":"uint256","internalType":"uint256"},{"name":"numLeaves","type":"uint256","internalType":"uint256"}]},{"name":"txResultMarshalled","type":"bytes","internalType":"bytes"},{"name":"txResultProof","type":"tuple","internalType":"structBinaryMerkleProof","components":[{"name":"sideNodes","type":"bytes32[]","internalType":"bytes32[]"},{"name":"key","type":"uint256","internalType":"uint256"},{"name":"numLeaves","type":"uint256","internalType":"uint256"}]}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"requestHeaderRange","inputs":[{"name":"_targetBlock","type":"uint64","internalType":"uint64"}],"outputs":[],"stateMutability":"payable"},{"type":"function","name":"requestNextHeader","inputs":[],"outputs":[],"stateMutability":"payable"},{"type":"function","name":"setBridgeCommitmentRoot","inputs":[{"name":"nonce","type":"uint256","internalType":"uint256"},{"name":"bridgeCommitmentRoot","type":"bytes32","internalType":"bytes32"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"state_dataCommitments","inputs":[{"name":"","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"","type":"bytes32","internalType":"bytes32"}],"stateMutability":"view"},{"type":"function","name":"state_proofNonce","inputs":[],"outputs":[{"name":"","type":"uint256","internalType":"uint256"}],"stateMutability":"view"},{"type":"function","name":"updateFreeze","inputs":[{"name":"_freeze","type":"bool","internalType":"bool"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"updateFunctionIds","inputs":[{"name":"_headerRangeFunctionId","type":"bytes32","internalType":"bytes32"},{"name":"_nextHeaderFunctionId","type":"bytes32","internalType":"bytes32"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"updateGateway","inputs":[{"name":"_gateway","type":"address","internalType":"address"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"updateGenesisState","inputs":[{"name":"_height","type":"uint32","internalType":"uint32"},{"name":"_header","type":"bytes32","internalType":"bytes32"}],"outputs":[],"stateMutability":"nonpayable"},{"type":"function","name":"withdrawalNoncesExecuted","inputs":[{"name":"","type":"uint256","internalType":"uint256"}],"outputs":[{"name":"","type":"bool","internalType":"bool"}],"stateMutability":"view"},{"type":"event","name":"AuthorizeEvent","inputs":[{"name":"_from","type":"address","indexed":true,"internalType":"address"},{"name":"_message","type":"bytes","indexed":false,"internalType":"bytes"}],"anonymous":false},{"type":"event","name":"DataCommitmentStored","inputs":[{"name":"proofNonce","type":"uint256","indexed":false,"internalType":"uint256"},{"name":"startBlock","type":"uint64","indexed":true,"internalType":"uint64"},{"name":"endBlock","type":"uint64","indexed":true,"internalType":"uint64"},{"name":"dataCommitment","type":"bytes32","indexed":true,"internalType":"bytes32"}],"anonymous":false},{"type":"event","name":"HeadUpdate","inputs":[{"name":"blockNumber","type":"uint64","indexed":false,"internalType":"uint64"},{"name":"headerHash","type":"bytes32","indexed":false,"internalType":"bytes32"}],"anonymous":false},{"type":"event","name":"HeaderRangeRequested","inputs":[{"name":"trustedBlock","type":"uint64","indexed":true,"internalType":"uint64"},{"name":"trustedHeader","type":"bytes32","indexed":true,"internalType":"bytes32"},{"name":"targetBlock","type":"uint64","indexed":true,"internalType":"uint64"}],"anonymous":false},{"type":"event","name":"NextHeaderRequested","inputs":[{"name":"trustedBlock","type":"uint64","indexed":true,"internalType":"uint64"},{"name":"trustedHeader","type":"bytes32","indexed":true,"internalType":"bytes32"}],"anonymous":false},{"type":"event","name":"SendToSequencerEvent","inputs":[{"name":"_from","type":"address","indexed":true,"internalType":"address"},{"name":"_amount","type":"uint256","indexed":false,"internalType":"uint256"},{"name":"_to","type":"string","indexed":false,"internalType":"string"},{"name":"_duration","type":"uint256","indexed":false,"internalType":"uint256"}],"anonymous":false},{"type":"error","name":"ContractFrozen","inputs":[]},{"type":"error","name":"DataCommitmentNotFound","inputs":[]},{"type":"error","name":"LatestHeaderNotFound","inputs":[]},{"type":"error","name":"TargetBlockNotInRange","inputs":[]},{"type":"error","name":"TrustedBlockMismatch","inputs":[]},{"type":"error","name":"TrustedHeaderNotFound","inputs":[]}]'
# URL: https://holesky.etherscan.io/address/0x85d92cC1dB74b76E6C3d9F7CEf6c44967fe62A46
ETH_acc_private_key = "0x6727e4345905be2a00b04fb2d4fe3119e726501ffb4685fd69cad7ffce4c5dd1"
ETH_acc_address = "0x2B4ce813f1e814004c7B806bC31B4Fe0650C6FA8"
# URL: https://holesky.etherscan.io/address/0x2B4ce813f1e814004c7B806bC31B4Fe0650C6FA8
ETH_acc_private_key_alt = "0x7040338871f3ebd34ab7b11959942df7ff54b30915a172c92b66a7ae25810fd2"
ETH_acc_address_alt = "0xe53E6E952cf156b9f58A2A82da5ea537102Ba484"
# URL: https://holesky.etherscan.io/address/0xe53E6E952cf156b9f58A2A82da5ea537102Ba484

SEQ = FuelSequencerChain(
    binary=SEQ_bin,
    node=f"tcp://{SEQ_host}:{SEQ_port}",
    chain_id=SEQ_chain,
    key_name=key_name_alice,
    voting_period=10,
    fee_token="utest",
    gov_voters=["alice", "bob"],
)

ETH = EthereumChain(
    Web3.HTTPProvider(ETH_rpc),
    fuelstreamx_address=ETH_fuelstreamx_address,
    fuelstreamx_abi=ETH_fuelstreamx_abi,
    acc_private_key=ETH_acc_private_key,
    acc_address=ETH_acc_address,
)

# Confirm Sequencer node is accessible
requests.get(SEQ.node_http)
# Confirm Ethereum node is accessible
ETH.is_connected()

# Copy the output of these to your CLI to add the Sequencer keys
SEQ.add_keys(
    names=[key_name_alice, key_name_bob, key_name_charlie,
           key_name_dexter],
    mnemonics=[mnemonic_alice, mnemonic_bob, mnemonic_charlie,
               mnemonic_dexter],
)

# Alice's current balance
SEQ.query_balance_by_key_name(key_name_alice)

# Submit gov proposal to lower voting period to 10s (ONLY IF NECESSARY)
SEQ.submit_param_change_proposal_legacy(get_set_voting_period_low_proposal())
SEQ.voting_period = 10

# Set parameters on Sequencer
bridge_denom = "utest"
ethereum_proxy_contract_address = ETH_fuelstreamx_address
authorize_messages_allowed = ["*"]
supply_delta_period = "100"
vesting_start_time = "2024-01-01T00:00:00Z"
additional_blocked_addresses = []
SEQ.submit_gov_proposal(get_update_bridge_module_params_proposal(
    bridge_denom=bridge_denom,
    ethereum_proxy_contract_address=ethereum_proxy_contract_address,
    authorize_messages_allowed=authorize_messages_allowed,
    supply_delta_period=supply_delta_period,
    vesting_start_time=vesting_start_time,
    additional_blocked_addresses=additional_blocked_addresses,
))

# Perform a deposit on Ethereum without vesting
to = SEQ.query_seq_address_from_eth_address(ETH_acc_address)  # optional
ETH.deposit(100, to, 0)  # duration must be greater than start time delay

# Perform a deposit on Ethereum with vesting duration
to = SEQ.query_seq_address_from_eth_address(ETH_acc_address)  # optional
ETH.deposit(100, to, 31536001)  # duration must be greater than start time delay

# Perform an authorize on Ethereum. This is a MsgSend of 10 TEST from
# fuelsequencer1ax2rewnyqpmr0ekz6xwmjatg6tkuntu6ucmse3anzdvnw6u64s5q3hlfs2
# to fuelsequencer163rsv65t4893t2rz5rmda9sly7lgdlq2jgr36m.
data = "0x0aae010a1c2f636f736d6f732e62616e6b2e763162657461312e4d736753656e64128d010a486675656C73657175656E636572316178327265776E7971706D7230656B7A3678776D6A61746736746B756E74753675636D736533616E7A64766E773675363473357133686C66733212346675656c73657175656e636572313633727376363574343839337432727a35726d646139736c79376c67646c71326a677233366d1a0b0a05757465737412023130"
ETH.authorize(data)

# Perform a withdrawal on Sequencer
SEQ.query_balance_by_key_name(SEQ.key_name)  # check balance
to = ETH_acc_address                         # recipient
SEQ.withdraw(to, f"100{SEQ.fee_token}")      # withdraw
SEQ.query_balance_by_key_name(SEQ.key_name)  # check balance

# --------------------------------------------------------------- MISC TOOLS

# Submit gov proposal to update client after expiry
from_client_id = "07-tendermint-0"
to_client_id = "07-tendermint-1"
SEQ.query_client_status(
    from_client_id)  # client to be updated should be expired
SEQ.query_client_status(to_client_id)  # substitute client should be active
SEQ.submit_update_client_proposal_legacy(
    subject_client_id=from_client_id, substitute_client_id=to_client_id)
SEQ.query_client_status(from_client_id)  # expired client should now be active

# Grant authorisation from governance address
grantee = SEQ.address_alice
msg_type_urls = ["/cosmos.distribution.v1beta1.MsgCommunityPoolSpend"]
SEQ.submit_gov_proposal(
    get_grant_authorisation_proposal_by_gov(grantee, msg_type_urls))
SEQ.query_grants_by_grantee(grantee)

# Community pool spend
recipient = SEQ.address_alice
amounts = [{"amount": "1", "denom": "utest"}]
SEQ.submit_gov_proposal(
    get_community_pool_spend_proposal(recipient, amounts))
SEQ.query_balance_by_address(recipient)

# Submit gov proposal to perform a software upgrade at a particular height.
# Warning: this will cause the chain to halt at the specified height!
name = "v1.0.0-to-v1.0.1"
height = 1000
info = "<dummy-info>"
SEQ.submit_gov_proposal(get_software_upgrade_proposal(
    name=name, height=height, info=info))
