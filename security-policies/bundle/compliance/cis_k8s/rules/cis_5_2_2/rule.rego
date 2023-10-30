package compliance.cis_k8s.rules.cis_5_2_2

import data.compliance.policy.kube_api.minimize_admission as audit

finding = result {
	result := audit.finding("privileged")
}
