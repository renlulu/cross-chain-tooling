# cross-chain-tooling

Command line tool used for migrating cross chain contracts.

## Commands

### deploy

```shell
go run ./main/main.go deploy -k <private_key>
```

Deploying cross chain manager contract, cross chain manager proxy contract and lock proxy contract. Also making proxy contract

pointing to cross chain manager. 

### migrate

```shell
go run ./main/main.go migrate -k <private_kay> -o <old_cross_chain_manager> -p <new_cross_chain_manager_proxy) -c <new_cross_chain_manager>
```

Migrating old cross chain manager's state to new cross chain manager contract.

#### Fields that will be migrated

* conKeepersPublicKeyList
* curEpochStartHeight
* zilToPolyTxHashMap (the latest one)
* zilToPolyTxHashIndex
* fromChainTxExist

Need determine how many we want to migrate for `zilToPolyTxHashMap` and `fromChainTxExist`.

#### Fields that will not be migrated

* whiteListFromContract
* whiteListToContract
* whiteListMethod

Need manually handle.

### Attention

After all migration job complete, we need change all admin account to multisig wallet.