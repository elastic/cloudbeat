package compliance.lib.output_validations

import data.compliance
import future.keywords.every

validate_metadata(metadata) {
	metadata.name
	metadata.description
	metadata.impact
	metadata.tags
	metadata.benchmark
	metadata.benchmark.name
	metadata.benchmark.version
	metadata.remediation
} else = false {
	true
}

# validate every rule metadata
test_validate_rule_metadata {
	all_rules := [rule | rule := compliance[benchmark].rules[rule_id]]

	every rule in all_rules {
		validate_metadata(rule.metadata)
	}
}
