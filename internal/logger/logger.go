package logger

import (
	"fmt"
	"log/slog"
	"os"
	"sync"

	"cicd/operators/helga/internal/vars"
	slogmulti "github.com/samber/slog-multi"
)

type logger_wrapper struct {
	logger   slog.Logger
	log_file os.File
}

var (
	logger_w *logger_wrapper
	once     sync.Once
)

func GetInstance() *logger_wrapper {
	once.Do(func() {
		logger_w = createLogger(vars.LOGS_FILE_PATH)
	})

	return logger_w
}

func GetLoggerInstance() *slog.Logger {
	return &GetInstance().logger
}

func createLogger(fname string) *logger_wrapper {
	log_f, err := os.OpenFile(fname, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("Failed to open log file: %v", err)
	}

	return &logger_wrapper{
		logger: *slog.New(
			slogmulti.Fanout(
				slog.NewJSONHandler(log_f, &slog.HandlerOptions{}),
				slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}),
			),
		),
		log_file: *log_f,
	}
}

// custom logger functions
func (l *logger_wrapper) CloseLogFile() {
	l.log_file.Close()
}
