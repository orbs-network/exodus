# Exodus

Exodus is a virtual chain migration tool.

It imports the block file into a local database, then loops over the table sending new transactions into the vchain. The tool can be stopped in any time and it will pick up where it left off because every transaction status is tracked in the database.

To enable the ability to you migrate the contract's state to a new vchain, you need to adjust your code to support it.

Please refer to [migration guide](MIGRATION-HOWTO.md) for refactoring tips.

## Running locally

```
docker run -d -p 5432:5432 -e POSTGRES_DB=exodus -e POSTGRES_USER=username -e POSTGRES_PASSWORD=password postgres
```

Deploy sample contract:

```
gamma-cli stop-local && gamma-cli start-local -override-config '{"consensus-context-maximum-transactions-in-block":1000,"transaction-pool-propagation-batch-size":500,"block-sync-num-blocks-in-batch":1000}' && gamma-cli deploy e2e/contract/notary.go -name NotaryV6 -signer user1
```

Update `config.json` file with relevant information.

To import transactions from `blocks` file into a local database, run

```
exodus import -config config.json
```

To start migration process to a new vchain, run

```
exodus migrate -config config.json
```
