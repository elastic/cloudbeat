package compliance.cis_eks

import data.compliance.cis_eks.rules

default_tags := ["CIS", "CIS v1.0.1", "EKS"]

benchmark_name := "CIS Amazon Elastic Kubernetes Service (EKS) Benchmark"

findings[finding] {
	# if activated rules were configured for this benchmark run only them
	data.activated_rules.cis_eks

	rule_id := data.activated_rules.cis_eks[_]
	finding = {
		"result": rules[rule_id].finding,
		"rule": rules[rule_id].metadata,
	}
}

findings[finding] {
	# no activated rules were configured for this benchmark, run all rules
	not data.activated_rules.cis_eks

	finding = {
		"result": rules[rule_id].finding,
		"rule": rules[rule_id].metadata,
	}
}
