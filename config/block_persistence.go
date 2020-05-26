package config

import (
	"github.com/orbs-network/orbs-spec/types/go/primitives"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
)

type BlockPersistenceConfig struct {
	Dir          string
	VirtualChain primitives.VirtualChainId
}

func (l *BlockPersistenceConfig) BlockStorageFileSystemDataDir() string {
	return l.Dir
}

func (l *BlockPersistenceConfig) BlockStorageFileSystemMaxBlockSizeInBytes() uint32 {
	return 64 * 1024 * 1024
}

func (l *BlockPersistenceConfig) VirtualChainId() primitives.VirtualChainId {
	return l.VirtualChain
}

func (l *BlockPersistenceConfig) NetworkType() protocol.SignerNetworkType {
	return protocol.SignerNetworkType(0)
}
