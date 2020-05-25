package db

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/orbs-network/orbs-client-sdk-go/orbs"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
)

var PUBLIC_KEY, _ = hex.DecodeString("B7Ef1A3E101322737416db57F7A2CC46DCc3Ae171870785CA072755638d2f1FF")
var PRIVATE_KEY, _ = hex.DecodeString("05E98d95c25815274679ff055aE49722CD5E2f888455AF392FDE2Bd3eBdB81B9B7Ef1a3E101322737416db57F7A2CC46DCC3ae171870785ca072755638d2F1Ff")

func Migrate(db *sql.DB, tableName string, client *orbs.OrbsClient, contractName string) error {
	rows, err := db.Query("SELECT * FROM "+tableName+" WHERE newTxStatus = $1 LIMIT $2 OFFSET $3", "", 10, 0)
	if err != nil {
		return err
	}

	dbTx, _ := db.Begin()

	for rows.Next() {
		var blockHeight uint64
		var timestamp uint64
		var rawArguments []byte
		var txId []byte
		var newTxIdPlaceholder []byte
		var newTxStatusPlaceholder string

		if err := rows.Scan(&blockHeight, &timestamp, &rawArguments, &txId, &newTxIdPlaceholder, &newTxStatusPlaceholder); err != nil {
			return err
		}

		inputArguments, err := protocol.PackedOutputArgumentsToNatives(rawArguments)
		if err != nil {
			return err
		}

		inputArgumentsWithTimestamp := append([]interface{}{timestamp}, inputArguments...)
		fmt.Println(inputArgumentsWithTimestamp)
		tx, newTxId, err := client.CreateTransaction(PUBLIC_KEY, PRIVATE_KEY, contractName, "importData",
			inputArgumentsWithTimestamp...)
		if err != nil {
			return err
		}
		res, err := client.SendTransaction(tx)
		if err != nil {
			fmt.Println(res.OutputArguments)
			return err
		}

		if _, err := db.Exec("UPDATE "+tableName+" SET newTxId = $1, newTxStatus = $2 WHERE txId = $3",
			newTxId, res.TransactionStatus.String(), txId); err != nil {
			fmt.Println(err)
			return dbTx.Rollback()
		}

		fmt.Println(res.ExecutionResult, res.TransactionStatus, res.OutputArguments)
	}

	return dbTx.Commit()
}
