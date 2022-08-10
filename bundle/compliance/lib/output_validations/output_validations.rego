package compliance.lib.output_validations

import data.compliance
import future.keywords.every

validate_common_kuberentes_provider_metadata(metadata) {
	metadata.id
	metadata.name
	metadata.profile_applicability
	metadata.description
	metadata.rationale
	metadata.audit
	metadata.remediation
	metadata.impact
	metadata.default_value
	metadata.references
	metadata.section
	metadata.version
	metadata.tags
	metadata.benchmark
	metadata.benchmark.name
	metadata.benchmark.version
	metadata.benchmark.id
}

validate_k8s_metadata(metadata) {
	validate_common_kuberentes_provider_metadata(metadata)
} else = false {
	true
}

validate_eks_metadata(metadata) {
	validate_common_kuberentes_provider_metadata(metadata)
} else = false {
	true
}

# validate every rule metadata
test_validate_rule_metadata {
	all_k8s_rules := [rule | rule := compliance.cis_k8s.rules[rule_id]]
	all_eks_rules := [rule | rule := compliance.cis_eks.rules[rule_id]]

	every k8s_rule in all_k8s_rules {
		validate_k8s_metadata(k8s_rule.metadata)
	}

	every eks_rule in all_eks_rules {
		validate_eks_metadata(eks_rule.metadata)
	}
}
