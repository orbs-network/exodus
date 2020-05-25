package main

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/orbs-network/crypto-lib-go/crypto/digest"
	dbImport "github.com/orbs-network/exodus/db"
	"github.com/orbs-network/orbs-client-sdk-go/codec"
	"github.com/orbs-network/orbs-client-sdk-go/orbs"
	"github.com/orbs-network/orbs-spec/types/go/primitives"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
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

type row struct {
	blockHeight uint64
	timestamp   uint64
	arguments   []byte
	txStatus    string
	txId        string
}

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

		if err := dbImport.Migrate(db, tableName, client, "NotaryV6"); err != nil {
			logger.Error("failure", log.Error(err))
		}

		return
	}

	// create table NotaryV1$register (blockHeight bigint, timestamp bigint, arguments bytea, txId varchar, newTxId varchar, newTxStatus varchar);

	if err := dbImport.Import(logger, &dbImport.BlockPersistenceConfig{
		ChainId: 1970000,
		Dir:     "/Users/kirill/Downloads/197",
	}, func(first primitives.BlockHeight, page []*protocol.BlockPairContainer) (wantsMore bool) {
		dbTx, _ := db.Begin()

		for _, block := range page {
			blockHeight := uint64(block.TransactionsBlock.Header.BlockHeight())
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

				if tx.ContractName() == contractName && tx.MethodName() == methodName {
					_, err := dbTx.Exec("INSERT INTO "+tableName+"(blockHeight, timestamp, arguments, txId, newTxId, newTxStatus) VALUES ($1, $2, $3, $4, $5, $6)",
						blockHeight, blockTimestamp, tx.RawInputArgumentArrayWithHeader(), txIdAsString, "", "")

					if err != nil {
						logger.Error("db error", log.Error(err))
						return false
					}
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
		return true
	}); err != nil {
		logger.Error("failed!", log.Error(err))
	} else {
		logger.Info("success!")
	}
}

func txIsSuccessfull(txHash []byte, receipts []*protocol.TransactionReceipt) bool {
	for _, txReceipt := range receipts {
		if bytes.Equal(txReceipt.Txhash(), txHash) {
			return txReceipt.ExecutionResult() == protocol.EXECUTION_RESULT_SUCCESS
		}
	}

	return false
}
