package embeded

import (
	"embed"
)

//go:embed experiments/*.yaml
var EmbeddedExperiments embed.FS

//go:embed components/*.yaml
var EmbeddedComponents embed.FS
