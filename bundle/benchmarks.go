package bundle

import "github.com/elastic/csp-security-policies/compliance"

func CISKubernetes() (map[string]string, error) {
	filePrefixes := []string{
		"main.rego",
		"lib/",
		"cis_k8s/",
	}

	return createPolicyMap(compliance.Embed, filePrefixes)
}

func CISEks() (map[string]string, error) {
	filePrefixes := []string{
		"main.rego",
		"lib/",
		"cis_eks/",
	}

	return createPolicyMap(compliance.Embed, filePrefixes)
}
