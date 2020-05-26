package db

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"github.com/orbs-network/crypto-lib-go/crypto/digest"
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

type ImportConfig struct {
	Contract    string
	Method      string
	BlockHeight primitives.BlockHeight // FIXME add min block height
	// FIXME add successful only
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

func Import(logger log.Logger, db *sql.DB, importConfig *ImportConfig, config *BlockPersistenceConfig) (error, int) {
	metricFactory := metric.NewRegistry()

	start := time.Now()
	persistence, err := filesystem.NewBlockPersistence(config, logger, metricFactory)

	logger.Info("startup time", log.String("duration", time.Since(start).String()))
	if err != nil {
		return err, 0
	}

	var count int
	if persistence.ScanBlocks(1, 255, func(first primitives.BlockHeight, page []*protocol.BlockPairContainer) (wantsMore bool) {
		dbTx, _ := db.Begin()

		var maxBlockHeight primitives.BlockHeight
		for _, block := range page {
			blockHeight := uint64(block.TransactionsBlock.Header.BlockHeight())
			blockTimestamp := uint64(block.TransactionsBlock.Header.Timestamp())

			maxBlockHeight = primitives.BlockHeight(blockHeight)

			for _, rawTx := range block.TransactionsBlock.SignedTransactions {
				tx := rawTx.Transaction()
				txHash := digest.CalcTxHash(tx)
				txId := digest.CalcTxId(tx)
				txIdAsString := "0x" + hex.EncodeToString(txId)

				// FIXME check txReceipt
				if !txIsSuccessfull(txHash, block.ResultsBlock.TransactionReceipts) {
					logger.Info("skipping tx", log.String("txId", txIdAsString))
					continue
				}

				if tx.ContractName() == importConfig.ContractName() && tx.MethodName() == importConfig.MethodName() {
					_, err := dbTx.Exec("INSERT INTO "+importConfig.TableName()+"(blockHeight, timestamp, arguments, txId, newTxId, newTxStatus) VALUES ($1, $2, $3, $4, $5, $6)",
						blockHeight, blockTimestamp, tx.RawInputArgumentArrayWithHeader(), txIdAsString, "", "")

					if err != nil {
						logger.Error("db error", log.Error(err))
						return false
					}

					count++
				}
			}

		}

		if err := dbTx.Commit(); err != nil {
			logger.Error("failed to commit to the db", log.Error(err))
			return false
		}

		logger.Info("processed block range",
			log.Uint64("start", uint64(page[0].TransactionsBlock.Header.BlockHeight())),
			log.Uint64("end", uint64(page[len(page)-1].TransactionsBlock.Header.BlockHeight())))

		return importConfig.MaxBlockHeight() > 0 && maxBlockHeight < importConfig.MaxBlockHeight()
	}); err != nil {
		return err, 0
	}

	return nil, count
}

func txIsSuccessfull(txHash []byte, receipts []*protocol.TransactionReceipt) bool {
	for _, txReceipt := range receipts {
		if bytes.Equal(txReceipt.Txhash(), txHash) {
			return txReceipt.ExecutionResult() == protocol.EXECUTION_RESULT_SUCCESS
		}
	}

	return false
}
