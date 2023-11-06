package compliance.cis_k8s.rules.cis_5_2_3

import data.compliance.policy.kube_api.minimize_sharing as audit

finding := audit.finding("hostPID")
