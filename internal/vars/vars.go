package vars

import (
	"os"
)

var (
	LOGS_FILE_PATH       = os.Getenv("LOGS_FILE_PATH")
	HELGA_CONF_FILE_PATH = os.Getenv("HELGA_CONF_FILE_PATH")
)