package compliance.lib.output_validations

import future.keywords.every

valid_metadata := {
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
	"benchmark": {
		"name": "benchmark",
		"version": "v1.0.0",
		"id": "cis_k8s",
	},
	"rule_number": "1.2.3",
}

test_required_metadata_fields {
	every key, _ in valid_metadata {
		not validate_metadata(object.remove(valid_metadata, [key]))
	}
}
