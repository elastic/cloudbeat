package compliance.lib.output_validations

###
### K8S Tests
###

test_validate_k8s_metadata_invalid_id {
	invalid_metadata := {
		"Id": "rule id", # <- capitalized. should be "id"
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
		"benchmark": {"name": "benchmark", "version": "v1.0.0"},
	}

	not validate_k8s_metadata(invalid_metadata)
}

test_validate_k8s_metadata_invalid_name {
	invalid_metadata := {
		"id": "rule id",
		"Name": "rule name", # <- capitalized. should be "name"
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
		"benchmark": {"name": "benchmark", "version": "v1.0.0"},
	}

	not validate_k8s_metadata(invalid_metadata)
}

test_validate_k8s_metadata_invalid_profile_applicability {
	invalid_metadata := {
		"id": "rule id",
		"name": "rule name",
		"Profile_applicability": "rule profile_applicability", # <- capitalized. should be "profile_applicability"
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
		"benchmark": {"name": "benchmark", "version": "v1.0.0"},
	}

	not validate_k8s_metadata(invalid_metadata)
}

test_validate_k8s_metadata_invalid_description {
	invalid_metadata := {
		"id": "rule id",
		"name": "rule name",
		"profile_applicability": "rule profile_applicability",
		"Description": "rule description", # <- capitalized. should be "description"
		"rationale": "rule rationale",
		"audit": "rule audit",
		"remediation": "rule remidiation",
		"impact": "rule impact",
		"default_value": "rule default_value",
		"references": "rule references",
		"section": "rule section",
		"version": "rule version",
		"tags": ["tag 1", "tag 2"],
		"benchmark": {"name": "benchmark", "version": "v1.0.0"},
	}

	not validate_k8s_metadata(invalid_metadata)
}

test_validate_k8s_metadata_invalid_rationale {
	invalid_metadata := {
		"id": "rule id",
		"name": "rule name",
		"profile_applicability": "rule profile_applicability",
		"description": "rule description",
		"Rationale": "rule rationale", # <- capitalized. should be "rationale"
		"audit": "rule audit",
		"remediation": "rule remidiation",
		"impact": "rule impact",
		"default_value": "rule default_value",
		"references": "rule references",
		"section": "rule section",
		"version": "rule version",
		"tags": ["tag 1", "tag 2"],
		"benchmark": {"name": "benchmark", "version": "v1.0.0"},
	}

	not validate_k8s_metadata(invalid_metadata)
}

test_validate_k8s_metadata_invalid_audit {
	invalid_metadata := {
		"id": "rule id",
		"name": "rule name",
		"profile_applicability": "rule profile_applicability",
		"description": "rule description",
		"rationale": "rule rationale",
		"Audit": "rule audit", # <- capitalized. should be "audit"
		"remediation": "rule remidiation",
		"impact": "rule impact",
		"default_value": "rule default_value",
		"references": "rule references",
		"section": "rule section",
		"version": "rule version",
		"tags": ["tag 1", "tag 2"],
		"benchmark": {"name": "benchmark", "version": "v1.0.0"},
	}

	not validate_k8s_metadata(invalid_metadata)
}

test_validate_k8s_metadata_invalid_remidiation {
	invalid_metadata := {
		"id": "rule id",
		"name": "rule name",
		"profile_applicability": "rule profile_applicability",
		"description": "rule description",
		"rationale": "rule rationale",
		"audit": "rule audit",
		"Remediation": "rule remidiation", # <- capitalized. should be "remidiation"
		"impact": "rule impact",
		"default_value": "rule default_value",
		"references": "rule references",
		"section": "rule section",
		"version": "rule version",
		"tags": ["tag 1", "tag 2"],
		"benchmark": {"name": "benchmark", "version": "v1.0.0"},
	}

	not validate_k8s_metadata(invalid_metadata)
}

test_validate_k8s_metadata_invalid_impact {
	invalid_metadata := {
		"id": "rule id",
		"name": "rule name",
		"profile_applicability": "rule profile_applicability",
		"description": "rule description",
		"rationale": "rule rationale",
		"audit": "rule audit",
		"remediation": "rule remidiation",
		"Impact": "rule impact", # <- capitalized. should be "impact"
		"default_value": "rule default_value",
		"references": "rule references",
		"section": "rule section",
		"version": "rule version",
		"tags": ["tag 1", "tag 2"],
		"benchmark": {"name": "benchmark", "version": "v1.0.0"},
	}

	not validate_k8s_metadata(invalid_metadata)
}

test_validate_k8s_metadata_invalid_default_value {
	invalid_metadata := {
		"id": "rule id",
		"name": "rule name",
		"profile_applicability": "rule profile_applicability",
		"description": "rule description",
		"rationale": "rule rationale",
		"audit": "rule audit",
		"remediation": "rule remidiation",
		"impact": "rule impact",
		"Default_value": "rule default_value", # <- capitalized. should be "default_value"
		"references": "rule references",
		"section": "rule section",
		"version": "rule version",
		"tags": ["tag 1", "tag 2"],
		"benchmark": {"name": "benchmark", "version": "v1.0.0"},
	}

	not validate_k8s_metadata(invalid_metadata)
}

test_validate_k8s_metadata_invalid_references {
	invalid_metadata := {
		"id": "rule id",
		"name": "rule name",
		"profile_applicability": "rule profile_applicability",
		"description": "rule description",
		"rationale": "rule rationale",
		"audit": "rule audit",
		"remediation": "rule remidiation",
		"impact": "rule impact",
		"default_value": "rule default_value",
		"References": "rule references", # <- capitalized. should be "references"
		"section": "rule section",
		"version": "rule version",
		"tags": ["tag 1", "tag 2"],
		"benchmark": {"name": "benchmark", "version": "v1.0.0"},
	}

	not validate_k8s_metadata(invalid_metadata)
}

test_validate_k8s_metadata_invalid_section {
	invalid_metadata := {
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
		"Section": "rule section", # <- capitalized. should be "section"
		"version": "rule version",
		"tags": ["tag 1", "tag 2"],
		"benchmark": {"name": "benchmark", "version": "v1.0.0"},
	}

	not validate_k8s_metadata(invalid_metadata)
}

test_validate_k8s_metadata_invalid_version {
	invalid_metadata := {
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
		"Version": "rule version", # <- capitalized. should be "version"
		"tags": ["tag 1", "tag 2"],
		"benchmark": {"name": "benchmark", "version": "v1.0.0"},
	}

	not validate_k8s_metadata(invalid_metadata)
}

test_validate_k8s_metadata_invalid_tags {
	invalid_metadata := {
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
		"Tags": ["tag 1", "tag 2"], # <- capitalized. should be "tags"
		"benchmark": {"name": "benchmark", "version": "v1.0.0"},
	}

	not validate_k8s_metadata(invalid_metadata)
}

test_validate_k8s_metadata_invalid_benchmark {
	invalid_metadata := {
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
		"Benchmark": {"name": "benchmark", "version": "v1.0.0"}, # <- capitalized. should be "benchmark"
	}

	not validate_k8s_metadata(invalid_metadata)
}

test_validate_k8s_metadata_valid {
	metadata := {
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
		"benchmark": {"name": "benchmark", "version": "v1.0.0"},
	}

	validate_k8s_metadata(metadata)
}

###
### EKS Tests
###

test_validate_eks_metadata_invalid_remediation {
	invalid_metadata := {
		"name": "rule name",
		"description": "rule description",
		"impact": "rule impact",
		"tags": ["tag 1", "tag 2"],
		"benchmark": {"name": "benchmark", "version": "v1.0.0"},
		"Remediation": "rule remidiation", # <- capitalized. should be "remediation"
	}

	not validate_eks_metadata(invalid_metadata)
}

test_validate_eks_metadata_invalid_remediation {
	invalid_metadata := {
		"name": "rule name",
		"description": "rule description",
		"impact": "rule impact",
		"tags": ["tag 1", "tag 2"],
		"benchmark": {"name": "benchmark", "version": "v1.0.0"},
		"Remediation": "rule remidiation", # <- capitalized. should be "remediation"
	}

	not validate_eks_metadata(invalid_metadata)
}

test_validate_eks_metadata_invalid_name {
	invalid_metadata := {
		"Name": "rule name", # <- capitalized. should be "name"
		"description": "rule description",
		"impact": "rule impact",
		"tags": ["tag 1", "tag 2"],
		"benchmark": {"name": "benchmark", "version": "v1.0.0"},
		"remediation": "rule remidiation",
	}

	not validate_eks_metadata(invalid_metadata)
}

test_validate_eks_metadata_invalid_desc {
	invalid_metadata := {
		"name": "rule name",
		"Description": "rule description", # <- capitalized. should be "Description"
		"impact": "rule impact",
		"tags": ["tag 1", "tag 2"],
		"benchmark": {"name": "benchmark", "version": "v1.0.0"},
		"remediation": "rule remidiation",
	}

	not validate_eks_metadata(invalid_metadata)
}

test_validate_eks_metadata_invalid_impact {
	invalid_metadata := {
		"name": "rule name",
		"description": "rule description",
		"Impact": "rule impact", # <- capitalized. should be "impact"
		"tags": ["tag 1", "tag 2"],
		"benchmark": {"name": "benchmark", "version": "v1.0.0"},
		"remediation": "rule remidiation",
	}

	not validate_eks_metadata(invalid_metadata)
}

test_validate_eks_metadata_invalid_tags {
	invalid_metadata := {
		"name": "rule name",
		"description": "rule description",
		"impact": "rule impact",
		"Tags": ["tag 1", "tag 2"], # <- capitalized. should be "tags"
		"benchmark": {"name": "benchmark", "version": "v1.0.0"},
		"remediation": "rule remidiation",
	}

	not validate_eks_metadata(invalid_metadata)
}

test_validate_eks_metadata_invalid_benchmark {
	invalid_metadata := {
		"name": "rule name",
		"description": "rule description",
		"impact": "rule impact",
		"tags": ["tag 1", "tag 2"],
		"Benchmark": {"name": "benchmark", "version": "v1.0.0"}, # <- capitalized. should be "benchmark"
		"remediation": "rule remidiation",
	}

	not validate_eks_metadata(invalid_metadata)
}

test_validate_eks_metadata_invalid_remediation {
	invalid_metadata := {
		"name": "rule name",
		"description": "rule description",
		"impact": "rule impact",
		"tags": ["tag 1", "tag 2"],
		"benchmark": {"name": "benchmark", "version": "v1.0.0"},
		"Remediation": "rule remidiation", # <- capitalized. should be "remediation"
	}

	not validate_eks_metadata(invalid_metadata)
}

test_validate_eks_metadata_valid {
	metadata := {
		"name": "rule name",
		"description": "rule description",
		"impact": "rule impact",
		"tags": ["tag 1", "tag 2"],
		"benchmark": {"name": "benchmark", "version": "v1.0.0"},
		"remediation": "rule remidiation",
	}

	validate_eks_metadata(metadata)
}
