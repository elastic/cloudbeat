package compliance.lib.output_validations

import future.keywords.every

k8s_valid_metadata := {
	"id": "rule id",
	"name": "rule name",
	"profile_applicability": "rule profile_applicability",
	"description": "rule description",
	"rationale": "rule rationale",
	"audit": "rule audit",
	"remediation": "rule remidiation",
	"impact": "rule impact",
	"default_value": "rule default_value",
	"references": "rule references",
	"section": "rule section",
	"version": "rule version",
	"tags": ["tag 1", "tag 2"],
	"benchmark": {"name": "benchmark", "version": "v1.0.0", "id": "cis_k8s"},
}

eks_valid_metadata := {
	"name": "rule name",
	"description": "rule description",
	"impact": "rule impact",
	"tags": ["tag 1", "tag 2"],
	"benchmark": {"name": "benchmark", "version": "v1.0.0", "id": "cis_eks"},
	"remediation": "rule remidiation",
}

test_required_metadata_fields {
	every key, _ in k8s_valid_metadata {
		not validate_k8s_metadata(object.remove(k8s_valid_metadata, [key]))
	}

	every key, _ in eks_valid_metadata {
		not validate_eks_metadata(object.remove(eks_valid_metadata, [key]))
	}
}
