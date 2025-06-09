package vars

import (
	"os"
)

var (
	LOGS_FILE_PATH       = os.Getenv("LOGS_FILE_PATH")
	HELGA_CONF_FILE_PATH = os.Getenv("HELGA_CONF_FILE_PATH")
	HOME                 = os.Getenv("HOME")
)

const (
	K8S_API_URL_REGEX               = `^(https?:\/\/)?[a-zA-Z0-9.-]+(:\d+)?$`
	ARTIFACTORY_VALIDATION_REGEX    = `^(https?:\/\/)?([a-zA-Z0-9-]+\.)*[a-zA-Z0-9-]+\/artifactory$`
	AQL_ARTIFACT_PATH_POSTFIX       = "api/search/aql"
	SYNC_INTERVAL_DEFAULT_RETENTION = 4
)
