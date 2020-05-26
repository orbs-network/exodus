package db

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"github.com/orbs-network/crypto-lib-go/crypto/digest"
	"github.com/orbs-network/exodus/config"
	"github.com/orbs-network/orbs-network-go/instrumentation/metric"
	"github.com/orbs-network/orbs-network-go/services/blockstorage/adapter/filesystem"
	"github.com/orbs-network/orbs-spec/types/go/primitives"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/orbs-network/scribe/log"
	"time"
)

func Import(logger log.Logger, db *sql.DB, importConfig *config.ImportConfig) (error, int) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS " + importConfig.TableName() + " (blockHeight bigint, timestamp bigint, arguments bytea, txId varchar, newTxId varchar, newTxStatus varchar)")
	if err != nil {
		return err, 0
	}

	metricFactory := metric.NewRegistry()

	start := time.Now()
	persistence, err := filesystem.NewBlockPersistence(&importConfig.BlockPersistence, logger, metricFactory)

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
