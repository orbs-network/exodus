package main

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	dbImport "github.com/orbs-network/exodus/db"
	"github.com/orbs-network/orbs-client-sdk-go/codec"
	"github.com/orbs-network/orbs-client-sdk-go/orbs"
	"github.com/orbs-network/scribe/log"
	"os"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "username"
	password = "password"
	dbname   = "exodus"

	contractName = "NotaryV1"
	methodName   = "register"

	tableName = contractName + "$" + methodName
)

var PUBLIC_KEY, _ = hex.DecodeString("B7Ef1A3E101322737416db57F7A2CC46DCc3Ae171870785CA072755638d2f1FF")
var PRIVATE_KEY, _ = hex.DecodeString("05E98d95c25815274679ff055aE49722CD5E2f888455AF392FDE2Bd3eBdB81B9B7Ef1a3E101322737416db57F7A2CC46DCC3ae171870785ca072755638d2F1Ff")

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	logger := log.GetLogger().WithOutput(log.NewFormattingOutput(os.Stdout, log.NewHumanReadableFormatter()))

	logger.Info("successfully connected to the database")

	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		client := orbs.NewClient("http://localhost:8080", 42, codec.NETWORK_TYPE_TEST_NET)

		if err := dbImport.Migrate(db, tableName, client, &orbs.OrbsAccount{
			PublicKey:  PUBLIC_KEY,
			PrivateKey: PRIVATE_KEY,
		}, "NotaryV6"); err != nil {
			logger.Error("failure", log.Error(err))
		}

		return
	}

	// create table NotaryV1$register (blockHeight bigint, timestamp bigint, arguments bytea, txId varchar, newTxId varchar, newTxStatus varchar);

	if err := dbImport.Import(logger, db, &dbImport.ImportConfig{
		Contract:    contractName,
		Method:      methodName,
		BlockHeight: 3000,
	}, &dbImport.BlockPersistenceConfig{
		ChainId: 1970000,
		Dir:     "/Users/kirill/Downloads/197",
	}); err != nil {
		logger.Error("failed!", log.Error(err))
	} else {
		logger.Info("success!")
	}
}
