package main

import (
	"flag"
	"github.com/orbs-network/exodus/actions"
	"github.com/orbs-network/exodus/config"
	"github.com/orbs-network/scribe/log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	logger := log.GetLogger().WithOutput(log.NewFormattingOutput(os.Stdout, log.NewHumanReadableFormatter()))
	cfgFlagSet := flag.NewFlagSet("config", flag.ExitOnError)

	if len(os.Args) > 1 {
		configPath := cfgFlagSet.String("config", "config.json", "path to config file")
		cfgFlagSet.Parse(os.Args[2:])

		logger.Info("loading config from", log.String("path", *configPath))
		cfg, err := config.GetConfig(*configPath)

		if err != nil {
			logger.Error("failed to parse the config file", log.Error(err))
			os.Exit(1)
		}

		start := time.Now()
		switch os.Args[1] {
		case "import":
			err = actions.Import(logger, cfg)
		case "migrate":
			err = actions.Migrate(logger, cfg)
		case "finish":
			err = actions.DisableImport(logger, cfg)
		}

		if err != nil {
			logger.Error("failure", log.Error(err), log.Stringable("duration", time.Since(start)))
			os.Exit(1)
		}

		logger.Info("success", log.Stringable("duration", time.Since(start)))
	} else {
		logger.Error("please enter a valid command")
		os.Exit(1)
	}
}
