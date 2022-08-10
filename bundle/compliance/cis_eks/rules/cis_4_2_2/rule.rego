package compliance.cis_eks.rules.cis_4_2_2

import data.compliance.policy.kube_api.minimize_sharing as audit

finding = result {
	result := audit.finding("hostPID")
}
