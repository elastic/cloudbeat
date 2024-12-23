package compliance.cis_k8s.rules.cis_1_3_6

import data.compliance.policy.process.ensure_arguments_contain_key_value as audit
import future.keywords.if

finding = result if {
	audit.controller_manager_filter
	result := audit.finding(audit.arg_contains("--feature-gates", "RotateKubeletServerCertificate=true"))
}
