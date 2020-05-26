package actions

import (
	"github.com/orbs-network/exodus/config"
	"github.com/orbs-network/exodus/db"
	"github.com/orbs-network/scribe/log"
)

func Import(logger log.Logger, cfg *config.Config) error {
	postgres, err := db.LocalConnection(cfg.Database.ConnectionString())
	if err != nil {
		return err
	}
	defer postgres.Close()

	if err, _ := db.Import(logger, postgres, &cfg.Import); err != nil {
		return err
	}

	return nil
}
