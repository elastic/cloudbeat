package compliance.lib.output_validations

test_validate_metadata_invalid_remediation {
	invalid_metadata := {
		"name": "rule name",
		"description": "rule description",
		"impact": "rule impact",
		"tags": ["tag 1", "tag 2"],
		"benchmark": {"name": "benchmark", "version": "v1.0.0"},
		"Remediation": "rule remidiation", # <- capitalized. should be "remediation"
	}

	not validate_metadata(invalid_metadata)
}

test_validate_metadata_invalid_name {
	invalid_metadata := {
		"Name": "rule name", # <- capitalized. should be "name"
		"description": "rule description",
		"impact": "rule impact",
		"tags": ["tag 1", "tag 2"],
		"benchmark": {"name": "benchmark", "version": "v1.0.0"},
		"remediation": "rule remidiation",
	}

	not validate_metadata(invalid_metadata)
}

test_validate_metadata_invalid_desc {
	invalid_metadata := {
		"name": "rule name",
		"Description": "rule description", # <- capitalized. should be "Description"
		"impact": "rule impact",
		"tags": ["tag 1", "tag 2"],
		"benchmark": {"name": "benchmark", "version": "v1.0.0"},
		"remediation": "rule remidiation",
	}

	not validate_metadata(invalid_metadata)
}

test_validate_metadata_invalid_impact {
	invalid_metadata := {
		"name": "rule name",
		"description": "rule description",
		"Impact": "rule impact", # <- capitalized. should be "impact"
		"tags": ["tag 1", "tag 2"],
		"benchmark": {"name": "benchmark", "version": "v1.0.0"},
		"remediation": "rule remidiation",
	}

	not validate_metadata(invalid_metadata)
}

test_validate_metadata_invalid_tags {
	invalid_metadata := {
		"name": "rule name",
		"description": "rule description",
		"impact": "rule impact",
		"Tags": ["tag 1", "tag 2"], # <- capitalized. should be "tags"
		"benchmark": {"name": "benchmark", "version": "v1.0.0"},
		"remediation": "rule remidiation",
	}

	not validate_metadata(invalid_metadata)
}

test_validate_metadata_invalid_benchmark {
	invalid_metadata := {
		"name": "rule name",
		"description": "rule description",
		"impact": "rule impact",
		"tags": ["tag 1", "tag 2"],
		"Benchmark": {"name": "benchmark", "version": "v1.0.0"}, # <- capitalized. should be "benchmark"
		"remediation": "rule remidiation",
	}

	not validate_metadata(invalid_metadata)
}

test_validate_metadata_invalid_remediation {
	invalid_metadata := {
		"name": "rule name",
		"description": "rule description",
		"impact": "rule impact",
		"tags": ["tag 1", "tag 2"],
		"benchmark": {"name": "benchmark", "version": "v1.0.0"},
		"Remediation": "rule remidiation", # <- capitalized. should be "remediation"
	}

	not validate_metadata(invalid_metadata)
}

test_validate_metadata_valid {
	metadata := {
		"name": "rule name",
		"description": "rule description",
		"impact": "rule impact",
		"tags": ["tag 1", "tag 2"],
		"benchmark": {"name": "benchmark", "version": "v1.0.0"},
		"remediation": "rule remidiation",
	}

	validate_metadata(metadata)
}
