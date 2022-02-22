package compliance.cis_k8s

import data.compliance.cis_k8s.rules

default_tags := ["CIS", "Kubernetes"]

benchmark_metadata := {
	"name": "CIS Kubernetes V1.20",
	"version": "v1.0.0",
}

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
	# if no activated rules were configured for this benchmark, run all rules
	not data.activated_rules.cis_k8s

	finding = {
		"result": rules[rule_id].finding,
		"rule": rules[rule_id].metadata,
	}
}
