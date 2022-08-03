package compliance.cis_k8s

import data.compliance.cis_k8s.rules

findings[finding] {
	# if activated rules were configured for this benchmark run only them
	input.activated_rules.cis_k8s

	rule_id := input.activated_rules.cis_k8s[_]
	finding = {
		"result": rules[rule_id].finding,
		"rule": rules[rule_id].metadata,
	}
}

findings[finding] {
	# if no activated rules were configured for this benchmark, run all rules
	not input.activated_rules.cis_k8s

	finding = {
		"result": rules[rule_id].finding,
		"rule": rules[rule_id].metadata,
	}
}
