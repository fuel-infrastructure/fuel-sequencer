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
    Networks.TESTNET: "https://testnet-rest-fuel-seq.simplystaking.xyz/cosmos/tx/v1beta1/txs/",
    Networks.SANDBOX: "https://sandbox-rest-fuel-seq.simplystaking.xyz/cosmos/tx/v1beta1/txs/",
    Networks.LOCAL: "http://localhost:1317/cosmos/tx/v1beta1/txs/",
}

SEQ_NODES = {
    Networks.TESTNET: "https://testnet-rpc-fuel-seq.simplystaking.xyz",
    Networks.SANDBOX: "https://sandbox-rpc-fuel-seq.simplystaking.xyz",
    Networks.LOCAL: "http://localhost:26657",
}

SEQ_RESTS = {
    Networks.TESTNET: "https://testnet-rest-fuel-seq.simplystaking.xyz",
    Networks.SANDBOX: "https://sandbox-rest-fuel-seq.simplystaking.xyz",
    Networks.LOCAL: "http://localhost:1317",
}

SEQ_CHAINS = {
    Networks.TESTNET: "seq-testnet-2",
    Networks.SANDBOX: "seq-sandbox-3",
    Networks.LOCAL: "fuelsequencer-1",
}

FEE_TOKENS = {
    Networks.TESTNET: "test",
    Networks.SANDBOX: "test",
    Networks.LOCAL: "utest",
}

GAS_PRICES = {
    Networks.TESTNET: f"10{FEE_TOKENS[Networks.TESTNET]}",
    Networks.SANDBOX: f"0{FEE_TOKENS[Networks.SANDBOX]}",
    Networks.LOCAL: f"0.025{FEE_TOKENS[Networks.LOCAL]}",
}
