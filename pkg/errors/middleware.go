package errors

import "github.com/fennet82/helga/internal/logger"

func HandleError(err error) {
	if err != nil {
		logger.GetLoggerInstance().Error(err.Error())
	}
}

func HandleErrors(errs []error) {
	for _, err := range errs {
		HandleError(err)
	}
}
