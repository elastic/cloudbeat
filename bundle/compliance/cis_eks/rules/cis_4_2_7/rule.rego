package compliance.cis_eks.rules.cis_4_2_7

import data.compliance.policy.kube_api.minimize_certain_capability as audit

finding = result {
	result := audit.finding
}
