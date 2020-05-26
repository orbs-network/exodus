package e2e

import (
	"database/sql"
	"fmt"
	"github.com/orbs-network/orbs-client-sdk-go/codec"
	"github.com/orbs-network/orbs-client-sdk-go/orbs"
	"github.com/orbs-network/orbs-contract-sdk/go/examples/test"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

type harness struct {
	client       *orbs.OrbsClient
	contractName string
	db           *sql.DB
}

func newHarness() *harness {
	return &harness{
		client:       orbs.NewClient(test.GetGammaEndpoint(), 42, codec.NETWORK_TYPE_TEST_NET),
		contractName: fmt.Sprintf("NotaryV%d", time.Now().UnixNano()),
	}
}

func (h *harness) deployContract(t *testing.T, sender *orbs.OrbsAccount) {
	contractSource, err := ioutil.ReadFile("./contract/notary.go")
	require.NoError(t, err)

	deployTx, _, err := h.client.CreateDeployTransaction(sender.PublicKey, sender.PrivateKey,
		h.contractName, orbs.PROCESSOR_TYPE_NATIVE, contractSource)
	require.NoError(t, err)

	deployResponse, err := h.client.SendTransaction(deployTx)
	require.NoError(t, err)

	require.EqualValues(t, codec.EXECUTION_RESULT_SUCCESS, deployResponse.ExecutionResult)
}

func (h *harness) dbConnect(t *testing.T) *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		"localhost", 5432, "username", "password", "exodus")

	db, err := sql.Open("postgres", psqlInfo)
	require.NoError(t, err)

	h.db = db

	return db
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
