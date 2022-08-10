package compliance.cis_k8s.rules.cis_5_2_10

import data.compliance.policy.kube_api.minimize_assigned_capabilities as audit

finding = result {
	result := audit.finding
}
