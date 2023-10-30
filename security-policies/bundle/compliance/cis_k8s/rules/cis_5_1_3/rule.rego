package compliance.cis_k8s.rules.cis_5_1_3

import data.compliance.policy.kube_api.minimize_wildcard as audit

finding = result {
	result := audit.finding
}
