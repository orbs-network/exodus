package config

import "github.com/orbs-network/orbs-spec/types/go/primitives"

type ImportConfig struct {
	Contract    string
	Method      string
	BlockHeight primitives.BlockHeight // FIXME add min block height
	// FIXME add successful only

	BlockPersistence BlockPersistenceConfig
}

func (c *ImportConfig) TableName() string {
	return c.Contract + "$" + c.Method
}

func (c *ImportConfig) ContractName() primitives.ContractName {
	return primitives.ContractName(c.Contract)
}

func (c *ImportConfig) MethodName() primitives.MethodName {
	return primitives.MethodName(c.Method)
}

func (c *ImportConfig) MaxBlockHeight() primitives.BlockHeight {
	return c.BlockHeight
}
