package compliance.cis_eks.rules.cis_4_2_4

import data.compliance.policy.kube_api.minimize_sharing as audit

finding := audit.finding("hostNetwork")
