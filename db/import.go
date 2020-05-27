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
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS " + importConfig.ContractName + " (blockHeight bigint, timestamp bigint, methodName varchar, arguments bytea, txId varchar, newTxId varchar, newTxStatus varchar)")
	if err != nil {
		return err, 0
	}

	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_" + importConfig.ContractName + "_txId ON " + importConfig.ContractName + "(txId)")
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

		wantsMore = true
		var lastBlockHeight primitives.BlockHeight
		for _, block := range page {
			lastBlockHeight = block.TransactionsBlock.Header.BlockHeight()
			blockHeight := uint64(lastBlockHeight)
			blockTimestamp := uint64(block.TransactionsBlock.Header.Timestamp())

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

				if tx.ContractName() == primitives.ContractName(importConfig.ContractName) {
					_, err := dbTx.Exec("INSERT INTO "+importConfig.ContractName+"(blockHeight, timestamp, methodName, arguments, txId, newTxId, newTxStatus) VALUES ($1, $2, $3, $4, $5, $6, $7)",
						blockHeight, blockTimestamp, tx.MethodName(), tx.RawInputArgumentArrayWithHeader(), txIdAsString, "", "")

					if err != nil {
						logger.Error("db error", log.Error(err))
						return false
					}

					count++
				}
			}

			if importConfig.BlockHeight > 0 && lastBlockHeight == importConfig.BlockHeight {
				wantsMore = false
				break
			}
		}

		if err := dbTx.Commit(); err != nil {
			logger.Error("failed to commit to the db", log.Error(err))
			return false
		}

		logger.Info("processed block range",
			log.Uint64("start", uint64(page[0].TransactionsBlock.Header.BlockHeight())),
			log.Uint64("end", uint64(lastBlockHeight)))

		return
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
