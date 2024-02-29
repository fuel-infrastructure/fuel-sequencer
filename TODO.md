ignite scaffold message SupplyDelta \
    --signer authority \
    --module bridge \
    --no-simulation -y

ignite scaffold message WithdrawToEthereum \
    nonce:string \
    to:string \
    amount:coin \
    --signer from \
    --module bridge \
    --no-simulation -y

ignite scaffold message PostBlob \
    nonce:string \
    topic:string \
    order:string \
    blob:string \
    --signer from \
    --module sequencing \
    --no-simulation -y

~~State, genesis, query: LastEthereumNonce~~ (OK)
~~State, genesis, query: LastEthereumBlockSynced~~ (OK)
State, genesis, query: EthereumEvents

State, genesis, query: Topics
~~State, genesis, query: MintAmount~~ (another PR)
~~State, genesis, query: BurnAmount~~ (another PR)

~~Param: VestingStartTime~~ (another PR)
~~Param: DepositContractAddress~~ (OK)
~~Param: AuthorizeContractAddress~~ (OK)
~~Param: AuthorizeMessagesAllowed~~ (OK)
~~Param: SupplyDeltaPeriod~~ (OK)
