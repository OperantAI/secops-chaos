package embedExperiments

import (
	"embed"
)

//go:embed *.yaml
var EmbeddedExperiments embed.FS
