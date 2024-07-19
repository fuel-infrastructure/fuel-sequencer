import json

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
    gov_voters=["<unused>"],
)
SEQ.gas_prices = f"10000000000{SEQ.fee_token}"

MIN_BALANCE = int(1 * 1e18)
FUND_AMOUNT = int(1e6 * 1e18)

ADDRESSES = [
    "fuelsequencer1ptrdx8rzykw57suy540tlhfhxfclmykywhtycf",
    "fuelsequencer1z6mj9pgyd0gdt7zpnn5k5xgw3u99ammylkh0tn",
    "fuelsequencer18q22y7cyxsev5ttqkvqt6dse0zd6zejyqja2h0",
    "fuelsequencer1wnfm34d2wees97cl34qcev8de3ja2n8whh732u",
    "fuelsequencer1h4dr6td79wagkgt5z8qhzaazh30nnavgq8g7wf",
    "fuelsequencer1w683zjxx9pvnceakafaf3penc060uhggdqjsw3",
    "fuelsequencer16eqy4fjtrr5rf8e0gfsc48qlzdpfl5enng849y",
    "fuelsequencer1ngsagnggmmumc62220n6waav5wzjgm8hf8kzh0",
    "fuelsequencer1cyvqtp695llvpcu08xg5ndjsqre0nsash972vz",
    "fuelsequencer1lut8dr0473pxm9ayynay7hhya9n8dnsa6egh5p",
]

for address in ADDRESSES:
    # Search for an existing balance
    bals = json.loads(SEQ.query_balance_by_address(address))['balances']
    bal = [b for b in bals if b['denom'] == SEQ.fee_token]
    if len(bal) == 1 and int(bal[0]['amount']) > MIN_BALANCE:
        print(f"Skipping {address}; already funded with {bal[0]['amount']}")
        continue

    SEQ.send(address, f"{FUND_AMOUNT}{SEQ.fee_token}")
