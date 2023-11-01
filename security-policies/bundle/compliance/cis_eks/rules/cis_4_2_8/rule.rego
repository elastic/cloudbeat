package compliance.cis_eks.rules.cis_4_2_8

import data.compliance.policy.kube_api.minimize_added_capabilities as audit

finding = result {
	result := audit.finding
}
