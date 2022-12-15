package config

// https://github.com/elastic/integrations/tree/main/packages/cloud_security_posture/data_stream/findings/agent/stream
const (
	CIS_K8S = "cis_k8s"
	CIS_EKS = "cis_eks"
	CIS_AWS = "cis_aws"
)

var SupportedCIS = []string{CIS_AWS, CIS_K8S, CIS_EKS}
