package compliance.cis_eks.rules.cis_4_2_5

import data.compliance.policy.kube_api.minimize_admission as audit

finding = result {
	result := audit.finding("allowPrivilegeEscalation")
}
