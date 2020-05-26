package e2e

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/orbs-network/exodus/config"
	"github.com/orbs-network/exodus/db"
	"github.com/orbs-network/orbs-client-sdk-go/codec"
	"github.com/orbs-network/orbs-client-sdk-go/orbs"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

type harness struct {
	client       *orbs.OrbsClient
	account      *orbs.OrbsAccount
	contractName string
	db           *sql.DB
}

func newHarness(t *testing.T, cfg *config.Config) *harness {
	db, err := db.LocalConnection(cfg.Database.ConnectionString())
	require.NoError(t, err)

	client := cfg.Orbs.Client()
	account, err := cfg.Orbs.Account()
	require.NoError(t, err)

	return &harness{
		client:       client,
		account:      account,
		db:           db,
		contractName: cfg.Orbs.Contract,
	}
}

func (h *harness) deployContract(t *testing.T) {
	contractSource, err := ioutil.ReadFile("./contract/notary.go")
	require.NoError(t, err)

	deployTx, _, err := h.client.CreateDeployTransaction(h.account.PublicKey, h.account.PrivateKey,
		h.contractName, orbs.PROCESSOR_TYPE_NATIVE, contractSource)
	require.NoError(t, err)

	deployResponse, err := h.client.SendTransaction(deployTx)
	require.NoError(t, err)

	require.EqualValues(t, codec.EXECUTION_RESULT_SUCCESS, deployResponse.ExecutionResult)
}

func (h *harness) dbTruncate(t *testing.T, tableName string) {
	h.db.Exec("TRUNCATE " + tableName)
}

func (h *harness) dbCountTransactions(t *testing.T, tableName string, status string) (count int) {
	result, err := h.db.Query("SELECT COUNT (*) FROM "+tableName+" WHERE newTxStatus = $1", status)
	require.NoError(t, err)

	result.Next()
	err = result.Scan(&count)
	require.NoError(t, err)

	return
}
