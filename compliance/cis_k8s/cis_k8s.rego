package compliance.cis_k8s

import data.compliance.cis_k8s.rules

default_tags := ["CIS", "CIS v1.6.0", "Kubernetes"]

benchmark_name := "CIS Kubernetes"

findings[finding] {
	# if activated rules were configured for this benchmark run only them
	data.activated_rules.cis_k8s

	rule_id := data.activated_rules.cis_k8s[_]
	finding = {
		"result": rules[rule_id].finding,
		"rule": rules[rule_id].metadata,
	}
}

findings[finding] {
	# no activated rules were configured for this benchmark, run all rules
	not data.activated_rules.cis_k8s

	finding = {
		"result": rules[rule_id].finding,
		"rule": rules[rule_id].metadata,
	}
}
