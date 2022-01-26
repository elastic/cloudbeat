package compliance.cis_eks

import data.compliance.cis_eks.rules

default_tags := ["CIS", "CIS v1.0.1", "EKS"]

benchmark_name := "CIS Amazon Elastic Kubernetes Service (EKS) Benchmark"

findings[finding] {
	some rule_id
	data.activated_rules.cis_eks[rule_id]
	finding = {
		"result": rules[rule_id].finding,
		"rule": rules[rule_id].metadata,
	}
}
