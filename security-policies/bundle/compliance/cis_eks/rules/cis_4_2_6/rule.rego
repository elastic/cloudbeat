package compliance.cis_eks.rules.cis_4_2_6

import data.compliance.policy.kube_api.minimize_admission_root as audit

finding = result {
	result := audit.finding
}
