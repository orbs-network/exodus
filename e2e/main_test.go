package e2e

import (
	"github.com/orbs-network/exodus/db"
	"github.com/orbs-network/orbs-client-sdk-go/codec"
	"github.com/orbs-network/orbs-client-sdk-go/orbs"
	"github.com/orbs-network/scribe/log"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestE2E(t *testing.T) {
	h := newHarness()

	account, _ := orbs.CreateAccount()
	h.deployContract(t, account)

	logger := log.GetLogger().WithOutput(log.NewFormattingOutput(os.Stdout, log.NewHumanReadableFormatter()))

	postgres := h.dbConnect(t)

	dbConfig := &db.ImportConfig{
		Contract:    "NotaryV1",
		Method:      "register",
		BlockHeight: 3000,
	}

	h.dbTruncate(t, dbConfig.TableName())

	err, importedTxCount := db.Import(logger, postgres, dbConfig, &db.BlockPersistenceConfig{
		ChainId: 1970000,
		Dir:     "/Users/kirill/Downloads/197",
	})
	require.NoError(t, err)
	require.Greater(t, importedTxCount, 0)

	txCount := h.dbCountTransactions(t, dbConfig.TableName(), "")
	require.EqualValues(t, txCount, importedTxCount)

	client := orbs.NewClient("http://localhost:8080", 42, codec.NETWORK_TYPE_TEST_NET)
	err, migratedTxCount := db.Migrate(logger, postgres, dbConfig.TableName(), client, account, h.contractName)
	require.NoError(t, err)

	require.EqualValues(t, migratedTxCount, h.dbCountTransactions(t, dbConfig.TableName(), "PENDING"))
	require.EqualValues(t, 0, h.dbCountTransactions(t, dbConfig.TableName(), "COMMITTED"))

	err = db.UpdateTxStatus(logger, postgres, dbConfig.TableName(), client, 300*time.Millisecond)
	require.NoError(t, err)

	require.EqualValues(t, 0, h.dbCountTransactions(t, dbConfig.TableName(), "PENDING"))
	require.EqualValues(t, migratedTxCount, h.dbCountTransactions(t, dbConfig.TableName(), "COMMITTED"))
}