package actions

import (
	"github.com/orbs-network/exodus/config"
	"github.com/orbs-network/exodus/db"
	"github.com/orbs-network/scribe/log"
	"time"
)

func Migrate(logger log.Logger, cfg *config.Config) error {
	for {
		postgres, err := db.LocalConnection(cfg.Database.ConnectionString())
		if err != nil {
			return err
		}

		logger.Info("connected to the database")

		start := time.Now()

		if err, count := db.Migrate(logger, postgres, cfg.Import.ContractName, cfg.Orbs); err != nil {
			postgres.Close()
			return err
		} else if count == 0 {
			break
		}

		logger.Info("transaction batch processed", log.Stringable("duration", time.Since(start)))
		postgres.Close()
	}

	return nil
}
