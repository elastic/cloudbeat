package compliance.cis_k8s.rules.cis_5_2_8

import data.compliance.policy.kube_api.minimize_certain_capability as audit

finding = result {
	result := audit.finding
}
