package compliance.lib.output_validations

import data.compliance

validate_metadata(metadata) {
	metadata.name
	metadata.description
	metadata.impact
	metadata.tags
	metadata.benchmark
	metadata.remediation
} else = false {
	true
}

# validate every rule metadata
test_validate_rule_metadata {
	all_rules := [rule | compliance[benchmark].rules[rule]]
	valid_rules := [rule | validate_metadata(compliance[benchmark].rules[rule].metadata)]

	count(valid_rules) == count(all_rules)
}
