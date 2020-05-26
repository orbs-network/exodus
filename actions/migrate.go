package actions

import (
	"github.com/orbs-network/exodus/config"
	"github.com/orbs-network/exodus/db"
	"github.com/orbs-network/scribe/log"
	"github.com/pkg/errors"
	"time"
)

func Migrate(logger log.Logger, cfg *config.Config) error {
	postgres, err := db.LocalConnection(cfg.Database.ConnectionString())
	if err != nil {
		return err
	}
	defer postgres.Close()

	client := cfg.Orbs.Client()
	account, err := cfg.Orbs.Account()
	if err != nil {
		errors.Wrap(err, "failed to access orbs account")

	}

	for {
		start := time.Now()

		if err, count := db.Migrate(logger, postgres, cfg.Import.TableName(), client, account, cfg.Orbs.Contract); err != nil {
			return err
		} else if count == 0 {
			break
		}

		if err := db.UpdateTxStatus(logger, postgres, cfg.Import.TableName(), client, 500*time.Millisecond); err != nil {
			return err
		}

		logger.Info("transaction batch processed", log.Stringable("duration", time.Since(start)))
	}

	return nil
}
