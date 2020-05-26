package main

import (
	"github.com/orbs-network/exodus/actions"
	"github.com/orbs-network/exodus/config"
	"github.com/orbs-network/scribe/log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	logger := log.GetLogger().WithOutput(log.NewFormattingOutput(os.Stdout, log.NewHumanReadableFormatter()))

	cfg, err := config.GetConfig("./config.json")
	if err != nil {
		logger.Error("failed to parse the config file", log.Error(err))
		os.Exit(1)
	}

	if len(os.Args) > 1 {

		switch os.Args[1] {
		case "import":
			err = actions.Import(logger, cfg)
		case "migrate":
			err = actions.Migrate(logger, cfg)
		}

		if err != nil {
			logger.Error("failure", log.Error(err))
		}
	}
}
