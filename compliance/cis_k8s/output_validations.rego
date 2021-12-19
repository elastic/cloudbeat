package compliance.cis_k8s.output_validations

import data.compliance.cis_k8s.rules
import data.compliance.lib.assert

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
	valid_rules := {rule_id | validate_metadata(rules[rule_id].metadata)}
	count(valid_rules) == count(rules)
}
