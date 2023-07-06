package compliance.lib.output_validations

import data.compliance
import future.keywords.every

validate_common_provider_metadata(metadata) {
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
	metadata.benchmark.rule_number
	metadata.benchmark.posture_type
}

validate_metadata(metadata) {
	validate_common_provider_metadata(metadata)
} else = false

# validate every rule metadata
test_validate_rule_metadata {
	all_k8s_rules := [rule | rule := compliance.cis_k8s.rules[rule_id]]
	all_eks_rules := [rule | rule := compliance.cis_eks.rules[rule_id]]
	all_aws_rules := [rule | rule := compliance.cis_aws.rules[rule_id]]
	all_gcp_rules := [rule | rule := compliance.cis_gcp.rules[rule_id]]

	print("Validating K8s rules")
	every k8s_rule in all_k8s_rules {
		validate_metadata(k8s_rule.metadata)
	}

	print("Validating EKS rules")
	every eks_rule in all_eks_rules {
		validate_metadata(eks_rule.metadata)
	}

	print("Validating AWS rules")
	every aws_rule in all_aws_rules {
		validate_metadata(aws_rule.metadata)
	}

	print("Validating GCP rules")
	every gcp_rule in all_gcp_rules {
		validate_metadata(gcp_rule.metadata)
	}
}
