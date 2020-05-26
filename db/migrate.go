package db

import (
	"database/sql"
	"github.com/orbs-network/orbs-client-sdk-go/orbs"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/orbs-network/scribe/log"
	"sync"
	"time"
)

type txIdPair struct {
	old string
	new string
}

type txIdError struct {
	id          string
	err         error
	description string
}

func Migrate(logger log.Logger, db *sql.DB, tableName string, client *orbs.OrbsClient, account *orbs.OrbsAccount, contractName string) (error, int) {
	// FIXME number of rows is not always the same as limit

	limit := 100
	offset := 0

	rows, err := db.Query("SELECT timestamp, arguments, txId FROM "+tableName+" WHERE newTxStatus = $1 LIMIT $2 OFFSET $3", "", limit, offset)
	if err != nil {
		return err, 0
	}

	reqStart := time.Now()

	dbTx, _ := db.Begin()
	var wg sync.WaitGroup
	errors := make(chan txIdError, limit)

	var count int
	for rows.Next() {
		wg.Add(1)

		var timestamp uint64
		var rawArguments []byte
		var txId string

		if err := rows.Scan(&timestamp, &rawArguments, &txId); err != nil {
			return err, 0
		}

		go func(rawArguments []byte, txId string) {
			defer wg.Done()

			<-time.After(1 * time.Microsecond)
			inputArguments, err := protocol.PackedOutputArgumentsToNatives(rawArguments)
			if err != nil {
				errors <- txIdError{
					id:  txId,
					err: err,
				}
				return
			}

			inputArgumentsWithTimestamp := append([]interface{}{timestamp}, inputArguments...)
			tx, newTxId, err := client.CreateTransaction(account.PublicKey, account.PrivateKey, contractName, "importData",
				inputArgumentsWithTimestamp...)
			if err != nil {
				logger.Error("failed to create new transaction", log.String("txId", txId))
				return
			}
			res, err := client.SendTransactionAsync(tx)
			if err != nil {
				var description string
				if res != nil && len(res.OutputArguments) > 0 {
					description = res.OutputArguments[0].(string)
				}

				logger.Error("failed to send new transaction", log.String("txId", txId), log.Error(err), log.String("remoteError", description))
				return
			}

			if _, err := dbTx.Exec("UPDATE "+tableName+" SET newTxId = $1, newTxStatus = $2 WHERE txId = $3",
				newTxId, res.TransactionStatus.String(), txId); err != nil {
				logger.Error("failed to update db", log.Error(err))
				return
			}
			count++
		}(rawArguments, txId)

	}

	wg.Wait()

	if err := dbTx.Commit(); err != nil {
		return err, 0
	}

	logger.Info("imported tx set", log.Int("total", count), log.Stringable("duration", time.Since(reqStart)))
	return nil, count
}
