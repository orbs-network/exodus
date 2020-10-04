package e2e

import (
	"fmt"
	"github.com/orbs-network/exodus/actions"
	"github.com/orbs-network/exodus/config"
	"github.com/orbs-network/exodus/db"
	"github.com/orbs-network/orbs-client-sdk-go/codec"
	"github.com/orbs-network/orbs-client-sdk-go/orbs"
	"github.com/orbs-network/scribe/log"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestE2E(t *testing.T) {
	cfg := &config.Config{
		Orbs: config.OrbsClientConfig{
			Endpoint:             "http://localhost:8080",
			VirtualChain:         42,
			PublicKey:            "0xb7ef1a3e101322737416db57f7a2cc46dcc3ae171870785ca072755638d2f1ff",
			PrivateKey:           "0x05e98d95c25815274679ff055ae49722cd5e2f888455af392fde2bd3ebdb81b9b7ef1a3e101322737416db57f7a2cc46dcc3ae171870785ca072755638d2f1ff",
			TransactionBatchSize: 100,
			ContractName:         fmt.Sprintf("NotaryV%d", time.Now().UnixNano()),
			ContractImportMethodMapping: map[string]string{
				"register": "importData",
			},
		},
		Database: config.DatabaseConfig{
			Database: "exodus",
			Host:     "localhost",
			Port:     5432,
			User:     "username",
			Password: "password",
		},
		Import: config.ImportConfig{
			ContractName: "NotaryV1",
			BlockHeight:  3000,
			BlockPersistence: config.BlockPersistenceConfig{
				VirtualChain: 1970000,
				Dir:          "/Users/kirill/Downloads/197",
			},
		},
	}

	h := newHarness(t, cfg)
	defer h.db.Close()

	h.deployContract(t)

	logger := log.GetLogger().WithOutput(log.NewFormattingOutput(os.Stdout, log.NewHumanReadableFormatter()))

	tableName := cfg.Import.ContractName
	h.dbTruncate(t, tableName)

	err, importedTxCount := db.Import(logger, h.db, &cfg.Import)
	require.NoError(t, err)
	require.Greater(t, importedTxCount, 0)

	txCount := h.dbCountTransactions(t, tableName, "")
	require.EqualValues(t, txCount, importedTxCount)

	err, migratedTxCount := db.Migrate(logger, h.db, tableName, cfg.Orbs)
	require.NoError(t, err)

	err = sendNotaryTransaction(cfg.Orbs.Client(), cfg.Orbs.ContractName)
	require.EqualError(t, err, "not allowed, data migration in progress")

	require.EqualValues(t, 0, h.dbCountTransactions(t, tableName, "PENDING"))
	require.EqualValues(t, migratedTxCount, h.dbCountTransactions(t, tableName, "COMMITTED"))

	err = actions.DisableImport(logger, cfg)
	require.NoError(t, err)

	err = sendNotaryTransaction(cfg.Orbs.Client(), cfg.Orbs.ContractName)
	require.NoError(t, err)
}

func sendNotaryTransaction(client *orbs.OrbsClient, contractName string) error {
	account, _ := orbs.CreateAccount()
	tx, _, _ := client.CreateTransaction(account.PublicKey, account.PrivateKey, contractName,
		"register", time.Now().String(), "", "")
	res, err := client.SendTransaction(tx)
	if err != nil {
		return err
	}

	if res.ExecutionResult == codec.EXECUTION_RESULT_SUCCESS && res.TransactionStatus == codec.TRANSACTION_STATUS_COMMITTED {
		return nil
	}

	return errors.New(res.OutputArguments[0].(string))
}
