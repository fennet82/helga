package main

import (
	"github.com/fennet82/helga/internal/logger"
	"github.com/fennet82/helga/pkg/config"
	helga_errors "github.com/fennet82/helga/pkg/errors"
)

var helgaConfig config.Config = config.Config{}

func main() {
	defer finalize()
	logger.GetLoggerInstance().Info("loading configuration...")

	errs := helgaConfig.UnmarshalYAMLConfig()
	if errs != nil {
		helga_errors.HandleErrors(errs)
		panic(errs)
	}

	logger.GetLoggerInstance().Info("configuration loaded successfully")
}

func finalize() {
	logger.GetLoggerInstance().Info("closing log file")
	logger.GetInstance().CloseLogFile()
}
