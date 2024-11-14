from utils.classes import CosmosChain
from utils.constants import *

host = "80.64.208.225"
port = "26657"
chain_id = "arbitrary"
bin = "fuelsequencerd"

chain = CosmosChain(
    binary=f"{bin}",
    node=f"tcp://{host}:{port}",
    chain_id=f"{chain_id}",
    key_name=key_name_alice,
    voting_period=1,
    fee_token="utest",
    gov_voters=["alice"],
)
chain.base64_events = False

height = 1  # or chain.get_last_block_height()
while True:
    block = chain.get_block(height)
    block_time = block['block']['header']['time']
    block_time = block_time.split('.')[0]  # simplify

    events = chain.get_block_events(height)
    for event in events["BB"]:
        print(f"{block_time} | height={height} | BB | {event}")
    for event in events["TX"]:
        print(f"{block_time} | height={height} | TX | {event}")
    for event in events["EB"]:
        print(f"{block_time} | height={height} | EB | {event}")
    height += 1
