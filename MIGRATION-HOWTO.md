# Refactoring smart contracts to support state migration to a new vchain

Migrating state between vchains is not an easy task, however, it is possible. Examples are available in `examples` directory.

First of all, let's examine a simple notary contract (which is not in any way fit for production):

```golang
package main

import (
	"github.com/orbs-network/orbs-contract-sdk/go/sdk/v1"
	"github.com/orbs-network/orbs-contract-sdk/go/sdk/v1/address"
	"github.com/orbs-network/orbs-contract-sdk/go/sdk/v1/env"
	"github.com/orbs-network/orbs-contract-sdk/go/sdk/v1/state"
)

var PUBLIC = sdk.Export(set, get)
var SYSTEM = sdk.Export(_init)

var OWNER_KEY = []byte("owner")

func _init() {
	// remember who deployed the contract
	state.WriteBytes(OWNER_KEY, address.GetSignerAddress())
}

func set(hash string) {
	state.WriteBytes([]byte(hash+"$owner"), address.GetSignerAddress())
	state.WriteUint64([]byte(hash+"$timestamp"), env.GetBlockTimestamp())
}

func get(hash string) (owner []byte, timestamp uint64) {
	owner = state.ReadBytes([]byte(hash + "$owner"))
	timestamp = state.ReadUint64([]byte(hash + "$timestamp"))
	return owner, timestamp
}
```

To migrate its state, we want to resend all the original transactions to the new vchain. However, vchain number is hardcoded in the transaction itself, and there is a cetain grace period (usually 30 minutes) within which the transaction will be admitted to the transaction pool. Original transaction are way too old (and invalid for the new chain) to be processed.

To mitigate this, we are going to create a bunch of new transactions and replace original `set` method with `importSet`, which will also incorporate the environmental data that is used in the original method: timestamp and the signer address.

Our new additional method will look like this:

```
func importSet(address []byte, timestamp timestamp, hash string) {
    state.WriteBytes([]byte(hash+"$owner"), address)
    state.WriteUint64([]byte(hash+"$timestamp"), timestamp)
}
```

To avoid code duplication, we should rewrite `set` and `importSet` to call the same base function:

```
func set(hash string) {
    _set(address.GetSignerAddress(), env.GetBlockTimestamp())
}

func importSet(address []byte, timestamp timestamp, hash string) {
    _set(address, timestamp, hash)
}

func _set(address []byte, timestamp timestamp, hash string) {
    state.WriteBytes([]byte(hash+"$owner"), address)
    state.WriteUint64([]byte(hash+"$timestamp"), timestamp)
}
```

Now our migration tool will be able to create new transactions that call `importSet(oldTxSigner, oldTxBlockTimestamp, hash)` to replicate the state between old contract and a new contract.

However, there is one thing missing: making sure that no one but the owner can call the import function (and that the owner can't do it after the migration was completed.

For that, we are going to introduce couple of new functions that we are going to call inside `importSet`:

```golang
func _ownerOnly() {
	if !bytes.Equal(state.ReadBytes(OWNER_KEY), address.GetSignerAddress()) {
		panic("not allowed!")
	}
}
```

`_ownerOnly` will revert any transaction that was sent from an address different from the contract owner address.

```golang
var DISABLE_IMPORT = []byte("DISABLE_IMPORT")

func disableImport() {
	_ownerOnly()
	state.WriteBool(DISABLE_IMPORT, true)
}

func _importAllowed() {
	if state.ReadBool(DISABLE_IMPORT) {
		panic("import is not allowed, data migration already finished")
	}
}

func _importDisabled() {
	if !state.ReadBool(DISABLE_IMPORT) {
		panic("not allowed, data migration in progress")
	}
}
```

`_importAllowed()` will revert every transaction if the import was not allowed (after the migration). The owner of the contract will have to call `disableImport` after all the data has been imported.

`_importDisabled()` will revert every transaction for anyone who tries to `set` anything when migration is in progress, effectively locking out everyone except the owner.

Finally, our new contract will look like this:

```golang
package main

import (
	"bytes"
	"github.com/orbs-network/orbs-contract-sdk/go/sdk/v1"
	"github.com/orbs-network/orbs-contract-sdk/go/sdk/v1/address"
	"github.com/orbs-network/orbs-contract-sdk/go/sdk/v1/env"
	"github.com/orbs-network/orbs-contract-sdk/go/sdk/v1/state"
)

var PUBLIC = sdk.Export(set, get, importSet, disableImport)
var SYSTEM = sdk.Export(_init)

var OWNER_KEY = []byte("owner")

func _init() {
	// remember who deployed the contract
	state.WriteBytes(OWNER_KEY, address.GetSignerAddress())
}

func set(hash string) {
    _importDisabled()
	_set(address.GetSignerAddress(), env.GetBlockTimestamp(), hash)
}

func get(hash string) (owner []byte, timestamp uint64) {
	owner = state.ReadBytes([]byte(hash + "$owner"))
	timestamp = state.ReadUint64([]byte(hash + "$timestamp"))
	return owner, timestamp
}

func importSet(address []byte, timestamp uint64, hash string) {
	_ownerOnly()
	_importAllowed()
	_set(address, timestamp, hash)
}

func _set(address []byte, timestamp uint64, hash string) {
	state.WriteBytes([]byte(hash+"$owner"), address)
	state.WriteUint64([]byte(hash+"$timestamp"), timestamp)
}

func _ownerOnly() {
	if !bytes.Equal(state.ReadBytes(OWNER_KEY), address.GetSignerAddress()) {
		panic("not allowed!")
	}
}

var DISABLE_IMPORT = []byte("DISABLE_IMPORT")

func disableImport() {
	_ownerOnly()
	state.WriteBool(DISABLE_IMPORT, true)
}

func _importAllowed() {
	if state.ReadBool(DISABLE_IMPORT) {
		panic("import is not allowed, data migration already finished")
	}
}

func _importDisabled() {
	if !state.ReadBool(DISABLE_IMPORT) {
		panic("not allowed, data migration in progress")
	}
}
```