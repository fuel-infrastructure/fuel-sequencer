from enum import Enum


class Networks(Enum):
    TESTNET = "TESTNET"
    SANDBOX = "SANDBOX"
    LOCAL = "LOCAL"


class NetworkConfig:

    def __init__(self, network: Networks):
        self.explorer_tx_url = EXPLORER_TX_URLS[network]

        self.seq_bin = "fuelsequencerd"  # needs to be in $GOPATH/bin
        self.seq_rpc = SEQ_NODES[network]
        self.seq_rest = SEQ_RESTS[network]
        self.seq_chain = SEQ_CHAINS[network]
        self.seq_fee_token = FEE_TOKENS[network]
        self.seq_gas_price = GAS_PRICES[network]


EXPLORER_TX_URLS = {
    Networks.TESTNET: "https://seq.simplystaking.xyz/fuel/tx/",
    Networks.SANDBOX: "http://80.64.208.225:1317/cosmos/tx/v1beta1/txs/",
    Networks.LOCAL: "http://localhost:1317/cosmos/tx/v1beta1/txs/",
}

SEQ_NODES = {
    Networks.TESTNET: "https://rpc-seq.simplystaking.xyz",
    Networks.SANDBOX: "http://80.64.208.225:26657",
    Networks.LOCAL: "http://localhost:26657",
}

SEQ_RESTS = {
    Networks.TESTNET: "https://rest-seq.simplystaking.xyz",
    Networks.SANDBOX: "http://80.64.208.225:1317",
    Networks.LOCAL: "http://localhost:1317",
}

SEQ_CHAINS = {
    Networks.TESTNET: "seq-testnet-1",
    Networks.SANDBOX: "seq-sandbox-2",
    Networks.LOCAL: "fuelsequencer-1",
}

FEE_TOKENS = {
    Networks.TESTNET: "utest",
    Networks.SANDBOX: "utest",
    Networks.LOCAL: "utest",
}

GAS_PRICES = {
    Networks.TESTNET: f"10000000000{FEE_TOKENS[Networks.TESTNET]}",
    Networks.SANDBOX: f"0.025{FEE_TOKENS[Networks.SANDBOX]}",
    Networks.LOCAL: f"0.025{FEE_TOKENS[Networks.LOCAL]}",
}
