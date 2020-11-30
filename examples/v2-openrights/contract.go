package main

import (
	"bytes"
	"github.com/orbs-network/orbs-contract-sdk/go/sdk/v1/env"
	"strings"

	"github.com/orbs-network/orbs-contract-sdk/go/sdk/v1/events"

	"github.com/orbs-network/orbs-contract-sdk/go/sdk/v1"
	"github.com/orbs-network/orbs-contract-sdk/go/sdk/v1/address"
	"github.com/orbs-network/orbs-contract-sdk/go/sdk/v1/state"
)

var SYSTEM = sdk.Export(_init)
var PUBLIC = sdk.Export(registerMedia, getMedia, importData, disableImport)
var EVENT = sdk.Export(mediaRegistered)

var OWNER_KEY = []byte("__CONTRACT_OWNER__")

func _init() {
	state.WriteBytes(OWNER_KEY, address.GetSignerAddress())
}

func isRegistered(pHash, binaryHash string) bool {
	panic("operation not supported, migrated to use events")

	key := []byte(pHash)
	records := state.ReadString(key)
	return strings.Contains(records, binaryHash)
}

func isValidURL(url string) bool {
	if len(url) == 0 {
		return false
	}
	if !strings.HasPrefix(url, "http") {
		return false
	}
	return true
}

func mediaRegistered(signer []byte, timestamp uint64, pHash, imageURL, postedAt, copyrights, binaryHash string) {
}

func validateInput(pHash, imageURL, postedAt, copyrights, binaryHash string) {
	if len(pHash) == 0 {
		panic("Image phash is not provided")
	}
	if !isValidURL(imageURL) {
		panic("Image URL is invalid")
	}
	if len(postedAt) == 0 {
		panic("Image timestamp is not provided")
	}
	if len(binaryHash) == 0 {
		panic("File image hash is not provided")
	}
}

func registerMedia(pHash, imageURL, postedAt, copyrights, binaryHash string) {
	if !bytes.Equal(state.ReadBytes(OWNER_KEY), address.GetSignerAddress()) {
		panic("Only contract owner can register media")
	}

	_register(address.GetSignerAddress(), env.GetBlockTimestamp(), pHash, imageURL, postedAt, copyrights, binaryHash)
}

func _register(signer []byte, ts uint64, pHash, imageURL, postedAt, copyrights, binaryHash string) {
	validateInput(pHash, imageURL, postedAt, copyrights, binaryHash)

	// skip
	//if isRegistered(pHash, binaryHash) {
	//	panic("Record with the following url already exists " + imageURL)
	//}

	//record := strings.Join([]string{imageURL, postedAt, copyrights, binaryHash}, ",")
	//key := []byte(pHash)
	//state.WriteString(key, state.ReadString(key)+"|"+record)

	events.EmitEvent(mediaRegistered, signer, ts, pHash, imageURL, postedAt, copyrights, binaryHash)
}

func getMedia(pHash string) []string {
	records := strings.TrimLeft(state.ReadString([]byte(pHash)), "|")
	return strings.Split(records, "|")
}

func _ownerOnly() {
	if !bytes.Equal(state.ReadBytes(OWNER_KEY), address.GetSignerAddress()) {
		panic("not allowed!")
	}
}

// Migration start

var DISABLE_IMPORT = []byte("DISABLE_IMPORT")
var METHODS_FOR_MIGRATION = []interface{}{importData, disableImport}

func importData(addr []byte, ts uint64, pHash, imageURL, postedAt, copyrights, binaryHash string) {
	_ownerOnly()
	_importAllowed()
	_register(addr, ts, pHash, imageURL, postedAt, copyrights, binaryHash)
}

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

// Migration end
