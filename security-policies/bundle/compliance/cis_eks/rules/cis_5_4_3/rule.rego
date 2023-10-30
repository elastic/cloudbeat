package compliance.cis_eks.rules.cis_5_4_3

import data.compliance.policy.kube_api.ensure_external_ip as audit

# Ensure there cluster node don't have a public IP
default rule_evaluation = true

# Verify that the node doesn't have an external IP
rule_evaluation = false {
	audit.verify_external_ip
}

# Ensure there cluster node don't have a public IP
finding = audit.finding(rule_evaluation)
