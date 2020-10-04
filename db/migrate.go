package db

import (
	"database/sql"
	"github.com/orbs-network/exodus/config"
	"github.com/orbs-network/orbs-client-sdk-go/orbs"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/orbs-network/scribe/log"
	"sync"
	"time"
)

func Status(db *sql.DB, contractName string) (err error, maxProcessedBlockHeight uint64) {
	rows, err := db.Query("SELECT  MAX(blockheight) FROM "+contractName+" WHERE newTxId != $1", "")
	if err != nil {
		return
	}

	if rows.Next() {
		err = rows.Scan(&maxProcessedBlockHeight)
	}

	return

}

func Migrate(logger log.Logger, db *sql.DB, contractName string, cfg config.OrbsClientConfig) (error, int) {
	account, err := cfg.Account()
	if err != nil {
		return err, 0
	}

	client := cfg.Client()

	rows, err := db.Query("SELECT timestamp, signer, methodName, arguments, txId FROM "+contractName+" WHERE newTxStatus = $1 ORDER BY blockHeight asc LIMIT $2", "", cfg.TransactionBatchSize)
	if err != nil {
		return err, 0
	}

	reqStart := time.Now()

	dbTx, _ := db.Begin()
	var wg sync.WaitGroup

	var count int
	for rows.Next() {
		wg.Add(1)

		var timestamp uint64
		var methodName string
		var signerRaw string
		var rawArguments []byte
		var txId string

		if err := rows.Scan(&timestamp, &signerRaw, &methodName, &rawArguments, &txId); err != nil {
			return err, 0
		}

		signer := orbs.AddressToBytes(signerRaw)

		go func(rawArguments []byte, txId string) {
			defer wg.Done()

			inputArguments, err := protocol.PackedOutputArgumentsToNatives(rawArguments)
			if err != nil {
				logger.Error("failed to parse tx input arguments", log.Error(err))
				return
			}

			inputArgumentsWithTimestamp := append([]interface{}{signer, timestamp}, inputArguments...)
			importMethodName := cfg.ContractImportMethodMapping[methodName]
			tx, newTxId, err := client.CreateTransaction(account.PublicKey, account.PrivateKey, cfg.ContractName, importMethodName,
				inputArgumentsWithTimestamp...)
			if err != nil {
				logger.Error("failed to create new transaction", log.String("txId", txId))
				return
			}
			res, err := client.SendTransaction(tx)
			if err != nil {
				var description string
				if res != nil && len(res.OutputArguments) > 0 {
					description = res.OutputArguments[0].(string)
				}

				logger.Error("failed to send new transaction", log.String("txId", txId), log.Error(err), log.String("remoteError", description))
				return
			}

			if _, err := dbTx.Exec("UPDATE "+contractName+" SET newTxId = $1, newTxStatus = $2 WHERE txId = $3",
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

	err, maxProcessedBlockHeight := Status(db, contractName)

	logger.Info("imported tx set", log.Int("total", count),
		log.Stringable("duration", time.Since(reqStart)),
		log.Uint64("processedBlockHeight", maxProcessedBlockHeight),
	)
	return nil, count
}
