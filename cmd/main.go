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

	err := helga_c.UnmarshalYAMLConfig()
	if err != nil {
		helga_errors.HandleError(err)

		return
	}

	finalizer()
}
