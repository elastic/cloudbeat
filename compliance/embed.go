package compliance

import "embed"

//go:embed lib
//go:embed cis_k8s
//go:embed main.rego
var Embed embed.FS
