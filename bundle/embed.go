package bundle

import "embed"

//go:embed compliance/cis_k8s
var CISRulesEmbed embed.FS

//go:embed compliance/cis_eks
var EKSRulesEmbed embed.FS

//go:embed compliance/main.rego
//go:embed compliance/lib
//go:embed compliance/kubernetes_common
var CommonEmbed embed.FS
