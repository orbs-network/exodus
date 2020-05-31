package actions

import (
	"fmt"
	"github.com/orbs-network/exodus/config"
	"github.com/orbs-network/orbs-client-sdk-go/codec"
	"github.com/orbs-network/scribe/log"
	"github.com/pkg/errors"
)

func DisableImport(logger log.Logger, cfg *config.Config) error {
	account, err := cfg.Orbs.Account()
	if err != nil {
		return err
	}

	client := cfg.Orbs.Client()

	tx, _, err := client.CreateTransaction(account.PublicKey, account.PrivateKey, cfg.Orbs.ContractName, "disableImport")
	if err != nil {
		return err
	}

	res, err := client.SendTransaction(tx)
	if err != nil {
		return err
	}

	if res.ExecutionResult == codec.EXECUTION_RESULT_SUCCESS && res.TransactionStatus == codec.TRANSACTION_STATUS_COMMITTED {
		return nil
	}

	if len(res.OutputArguments) > 0 {
		return errors.New(fmt.Sprintf("failed to disable import : %s", res.OutputArguments[0].(string)))
	}

	return errors.New("failed to disable import")
}
