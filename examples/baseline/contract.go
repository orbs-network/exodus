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
