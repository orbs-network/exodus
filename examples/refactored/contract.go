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
