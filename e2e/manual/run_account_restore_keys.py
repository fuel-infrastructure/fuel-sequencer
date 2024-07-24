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

MNEMONICS = [
    "crunch clap original cinnamon evolve tail release media first cabin surge jar predict pluck ski fiber never happy swing scrub stadium shaft shrimp switch",
    "mutual tide they soup wave corn pulse galaxy drama trap boss blanket stuff tribe citizen remove lift inside neither shock monkey daughter toss food",
    "coil multiply author weasel snake winter cereal garment guide update shaft hover detail remove cannon suffer write auction gun reveal excite raccoon divorce world",
    "imitate parent start glide gather various visit inform cram ill number capital trick seat useless swap grief anchor cart giggle electric broken favorite price",
    "chase guide clerk like frown armor upset arrest flavor error loop address resist duty appear century disorder fence list fog physical castle erode shuffle",
    "source fatal unveil angry party enhance stock desert rocket valve rough aisle snack truth dove inspire claw latin donkey puzzle slim current organ canyon",
    "auto ocean close when intact unfair quick brick sport mass expand month arrest finger crunch bring tag canoe ring relief drum blame shield model",
    "home pioneer dose draft lucky program carry pepper spot control elephant enjoy foster vacuum shoot trial april between upon nephew dignity mass treat will",
    "large east file require guard ahead margin rabbit since omit margin stable whale balance various when swamp shift people merit explain addict struggle fury",
    "intact frog acquire grow boat diagram cost people limit emotion develop life bullet maze gospel physical custom prize undo symbol swallow announce lonely mean",
]

KEY_NAMES = [
    "temp1",
    "temp2",
    "temp3",
    "temp4",
    "temp5",
    "temp6",
    "temp7",
    "temp8",
    "temp9",
    "temp10",
]

SEQ.add_keys(names=KEY_NAMES, mnemonics=MNEMONICS)

for key in KEY_NAMES:
    print(f"{key} :: {SEQ.keys(f'show {key} -a')}")
