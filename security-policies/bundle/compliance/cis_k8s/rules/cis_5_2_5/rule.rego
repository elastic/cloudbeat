package compliance.cis_k8s.rules.cis_5_2_5

import data.compliance.policy.kube_api.minimize_sharing as audit

finding = result {
	result := audit.finding("hostNetwork")
}
