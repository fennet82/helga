package vars

import (
	"os"
)

var (
	LOGS_FILE_PATH       = os.Getenv("LOGS_FILE_PATH")
	HELGA_CONF_FILE_PATH = os.Getenv("HELGA_CONF_FILE_PATH")
)

const (
	DOMAIN_VALIDATION_REGEX = `^(https?:\/\/)?[a-zA-Z0-9][a-zA-Z0-9.-]*(:[0-9]+)?$`
	AQL_ARTIFACT_PATH_POSTFIX = "/artifactory/api/search/aql"
)