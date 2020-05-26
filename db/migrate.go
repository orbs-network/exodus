package db

import (
	"database/sql"
	"fmt"
	"github.com/orbs-network/orbs-client-sdk-go/codec"
	"github.com/orbs-network/orbs-client-sdk-go/orbs"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
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

func Migrate(db *sql.DB, tableName string, client *orbs.OrbsClient, account *orbs.OrbsAccount, contractName string) error {
	// FIXME number of rows is not always the same as limit

	limit := 200
	offset := 0

	rows, err := db.Query("SELECT timestamp, arguments, txId FROM "+tableName+" WHERE newTxStatus = $1 LIMIT $2 OFFSET $3", "", limit, offset)
	if err != nil {
		return err
	}

	reqStart := time.Now()

	dbTx, _ := db.Begin()
	var wg sync.WaitGroup
	txIdPairs := make(chan txIdPair, limit)
	errors := make(chan txIdError, limit)

	for rows.Next() {
		wg.Add(1)

		var timestamp uint64
		var rawArguments []byte
		var txId string

		if err := rows.Scan(&timestamp, &rawArguments, &txId); err != nil {
			return err
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
			fmt.Println(inputArgumentsWithTimestamp)
			tx, newTxId, err := client.CreateTransaction(account.PublicKey, account.PrivateKey, contractName, "importData",
				inputArgumentsWithTimestamp...)
			if err != nil {
				errors <- txIdError{
					id:  txId,
					err: err,
				}
				return
			}
			res, err := client.SendTransactionAsync(tx)
			if err != nil {
				var description string
				if res != nil && len(res.OutputArguments) > 0 {
					description = res.OutputArguments[0].(string)
				}

				errors <- txIdError{
					id:          txId,
					err:         err,
					description: description,
				}
				return
			}

			if _, err := db.Exec("UPDATE "+tableName+" SET newTxId = $1, newTxStatus = $2 WHERE txId = $3",
				newTxId, res.TransactionStatus.String(), txId); err != nil {
				fmt.Println(err)
				errors <- txIdError{
					id:  txId,
					err: err,
				}
				return
			}

			txIdPairs <- txIdPair{
				old: txId,
				new: newTxId,
			}
		}(rawArguments, txId)

	}

	wg.Wait()

	dbTx.Commit()

	fmt.Println(time.Since(reqStart))

	dbTx, _ = db.Begin()

	statusStart := time.Now()
	success := 0

	totalPairs := len(txIdPairs)
	for i := 0; i < totalPairs; i++ {
		select {
		case pair := <-txIdPairs:
			wg.Add(1)
			go func(pair txIdPair) {
				defer wg.Done()

				for {
					<-time.After(100 * time.Millisecond)

					res, err := client.GetTransactionStatus(pair.new)
					if err == nil {
						fmt.Println("TX STATUS:", pair.new, res.TransactionStatus)
					}

					if err != nil || res.TransactionStatus == codec.TRANSACTION_STATUS_PENDING {
						continue
					}

					json, _ := res.MarshalJSON()
					fmt.Println(string(json))

					fmt.Println("UPDATE TO COMPLETE:", pair.new)

					if _, err := db.Exec("UPDATE "+tableName+" SET newTxStatus = $1 WHERE newTxId = $2",
						res.TransactionStatus.String(), pair.new); err != nil {

						//errors <- txIdError{
						//	id:  txId,
						//	err: err,
						//}
						fmt.Println("ERR", err)
						return
					} else {
						fmt.Println("BREAK", pair.new, res.TransactionStatus)
						success++
						break
					}
				}
			}(pair)
		}
	}

	wg.Wait()

	fmt.Println(success, "/", totalPairs)

	if err := dbTx.Commit(); err != nil {
		fmt.Println("ERR", err)
	}

	fmt.Println(time.Since(statusStart))

	return nil
}
