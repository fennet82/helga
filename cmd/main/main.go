package main

import (
	"sync"

	"github.com/fennet82/helga/internal/logger"
	"github.com/fennet82/helga/pkg/config"
	helga_errors "github.com/fennet82/helga/pkg/errors"
)

var helgaConfig config.Config = config.Config{}

func main() {
	var wg sync.WaitGroup

	defer finalize()
	logger.GetLoggerInstance().Info("loading configuration...")

	errs := helgaConfig.UnmarshalYAMLConfig()
	if errs != nil {
		helga_errors.HandleErrors(errs)
		panic(errs)
	}

	logger.GetLoggerInstance().Info("configuration loaded successfully")
	logger.GetLoggerInstance().Info("starting to initiate clusters")

	for _, c := range helgaConfig.Clusters {
		c.Init()

		wg.Add(1)
		go func() {
			defer wg.Done()
			c.SyncNamespacesWithCluster()
		}()
	}
	wg.Wait()

	logger.GetLoggerInstance().Info("finished.")
}

func finalize() {
	logger.GetLoggerInstance().Info("closing log file")
	logger.GetInstance().CloseLogFile()
}
