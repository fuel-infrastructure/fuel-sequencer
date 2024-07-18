mnemonic_alice = "dinner crash nurse casino baby fold race cheese elite column sausage sleep close royal rain over mechanic minimum outdoor conduct cash wagon frog evidence"
mnemonic_bob = "gaze drama excess raven follow antenna swallow beef upper myself question pitch course ill adult century crisp ice rough match praise sing unveil vintage"
mnemonic_charlie = "bar describe panda mosquito quiz room daring round nurse disagree swallow frown hat repeat recall flight skin sketch volume dutch range grunt assist nerve"
mnemonic_dexter = "bonus clinic owner choose grief soda ride divorce album oval tone mixed mechanic coin defense wonder tumble vault sorry great hover neither security amazing"

key_name_alice = "alice"
key_name_bob = "bob"
key_name_charlie = "charlie"
key_name_dexter = "dexter"
key_name_eve = "eve"
key_name_proposer = "proposer"

events_filter = [
    "coin_spent", "coin_received", "transfer", "mint", "coinbase",

    "tx", "message",

    "rewards", "commission", "proposer_reward",

    "create_client", "update_client",

    "ibc_transfer", "fungible_token_packet",

    "connection_open_init", "connection_open_ack",
    "channel_open_init", "channel_open_ack",

    "proposal_vote", "proposal_deposit",

    "execute", "wasm", "reply",  # wasm
    "liveness",
]

events_filter_by_prefix = [
    "wasm-"
]
