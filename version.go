package s3ctl

import _ "embed"

//go:embed .version
var version string

func Version() string {
	return version
}
