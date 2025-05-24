package main

import (
	"cicd/operators/helga/config"
	helga_errors "cicd/operators/helga/errors"
	"cicd/operators/helga/internal/logger"
)

func main() {
	var helga_c config.Config

	logger.GetLoggerInstance().Info("loading configuration...")

	errs := helga_c.UnmarshalYAMLConfig()
	if errs != nil {
		helga_errors.HandleErrors(errs)
		panic(errs)
	}

	defer finalize()
}

func finalize() {
	logger.GetLoggerInstance().Error("closing log file")
	logger.GetInstance().CloseLogFile()
}
