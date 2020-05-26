package config

import "github.com/orbs-network/orbs-spec/types/go/primitives"

type ImportConfig struct {
	ContractName string
	BlockHeight  primitives.BlockHeight // FIXME add min block height
	// FIXME add successful only

	BlockPersistence BlockPersistenceConfig
}
