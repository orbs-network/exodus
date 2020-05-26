package db

import (
	"database/sql"
	"github.com/orbs-network/exodus/config"
	"github.com/orbs-network/orbs-client-sdk-go/codec"
	"github.com/orbs-network/scribe/log"
	"sync"
	"time"
)

func UpdateTxStatus(logger log.Logger, db *sql.DB, tableName string, cfg config.OrbsClientConfig) error {
	client := cfg.Client()
	rows, err := db.Query("SELECT txId, newTxId FROM "+tableName+" WHERE newTxStatus = $1 LIMIT $2", "PENDING", cfg.TransactionBatchSize)
	if err != nil {
		return err
	}

	var txIdPairs []txIdPair
	for rows.Next() {
		var old string
		var new string

		if err := rows.Scan(&old, &new); err != nil {
			return err
		}

		txIdPairs = append(txIdPairs, txIdPair{old, new})
	}

	var wg sync.WaitGroup

	dbTx, _ := db.Begin()

	statusStart := time.Now()
	success := 0

	totalPairs := len(txIdPairs)
	for _, pair := range txIdPairs {
		wg.Add(1)
		go func(pair txIdPair) {
			defer wg.Done()

			for {
				<-time.After(time.Duration(cfg.TransactionStatusQueryIntervalInMs) * time.Millisecond)

				res, err := client.GetTransactionStatus(pair.new)
				if err != nil || res.TransactionStatus == codec.TRANSACTION_STATUS_PENDING {
					continue
				}

				if _, err := dbTx.Exec("UPDATE "+tableName+" SET newTxStatus = $1 WHERE newTxId = $2",
					res.TransactionStatus.String(), pair.new); err != nil {

					logger.Error("failed to update db", log.Error(err))
					return
				} else {
					logger.Error("update tx status", log.String("txId", pair.old),
						log.String("newTxId", pair.new), log.Stringable("status", res.TransactionStatus))
					success++
					break
				}
			}
		}(pair)
	}

	wg.Wait()

	if err := dbTx.Commit(); err != nil {
		return err
	}

	logger.Info("processed tx set", log.Int("successful", success), log.Int("total", totalPairs), log.Stringable("duration", time.Since(statusStart)))

	return nil
}
