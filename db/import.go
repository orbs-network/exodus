package db

import (
	"github.com/orbs-network/orbs-network-go/instrumentation/metric"
	"github.com/orbs-network/orbs-network-go/services/blockstorage/adapter/filesystem"
	"github.com/orbs-network/orbs-spec/types/go/primitives"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/orbs-network/scribe/log"
	"time"
)

type BlockPersistenceConfig struct {
	Dir     string
	ChainId primitives.VirtualChainId
}

func (l *BlockPersistenceConfig) BlockStorageFileSystemDataDir() string {
	return l.Dir
}

func (l *BlockPersistenceConfig) BlockStorageFileSystemMaxBlockSizeInBytes() uint32 {
	return 64 * 1024 * 1024
}

func (l *BlockPersistenceConfig) VirtualChainId() primitives.VirtualChainId {
	return l.ChainId
}

func (l *BlockPersistenceConfig) NetworkType() protocol.SignerNetworkType {
	return protocol.SignerNetworkType(0)
}

func Import(logger log.Logger, config *BlockPersistenceConfig,
	callback func(first primitives.BlockHeight, page []*protocol.BlockPairContainer) (wantsMore bool)) error {
	metricFactory := metric.NewRegistry()

	start := time.Now()
	persistence, err := filesystem.NewBlockPersistence(config, logger, metricFactory)

	logger.Info("startup time", log.String("duration", time.Since(start).String()))
	if err != nil {
		return err
	}

	if err := persistence.ScanBlocks(1, 255, callback); err != nil {
		return err
	}

	return nil
}
