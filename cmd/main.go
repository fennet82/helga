package main

import (
	"cicd/operators/helga/config"
	helga_errors "cicd/operators/helga/errors"
	"cicd/operators/helga/internal/logger"
)

func finalizer() {
	logger.GetInstance().CloseLogFile()
}

func main() {
	var helga_c config.Config

	errs := helga_c.UnmarshalYAMLConfig()
	if errs != nil {
		helga_errors.HandleErrors(errs)

		return
	}

	finalizer()
}
