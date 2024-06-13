package embedComponents

import (
	"embed"
)

//go:embed *.yaml
var EmbeddedComponents embed.FS
