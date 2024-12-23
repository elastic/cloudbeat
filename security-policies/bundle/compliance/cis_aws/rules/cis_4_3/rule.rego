package compliance.cis_aws.rules.cis_4_3

import data.compliance.lib.common
import data.compliance.policy.aws_cloudtrail.data_adapter
import data.compliance.policy.aws_cloudtrail.pattern
import data.compliance.policy.aws_cloudtrail.trail
import future.keywords.if

default rule_evaluation := false

finding := result if {
	# filter
	data_adapter.is_multi_trails_type

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(rule_evaluation),
		input.resource,
	)
}

# { $.userIdentity.type = \"Root\" && $.userIdentity.invokedBy NOT EXISTS && $.eventType != \"AwsServiceEvent\" }
required_patterns := [pattern.complex_expression("&&", [
	pattern.simple_expression("$.userIdentity.type", "=", "\"Root\""),
	pattern.simple_expression("$.userIdentity.invokedBy", "NOT EXISTS", ""),
	pattern.simple_expression("$.eventType", "!=", "\"AwsServiceEvent\""),
])]

rule_evaluation := trail.at_least_one_trail_satisfied(required_patterns)
