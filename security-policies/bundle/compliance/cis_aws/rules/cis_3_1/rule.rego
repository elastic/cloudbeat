package compliance.cis_aws.rules.cis_3_1

import data.compliance.lib.common
import data.compliance.policy.aws_cloudtrail.data_adapter
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

rule_evaluation if {
	some i
	t := data_adapter.trail_items[i]
	trail.cloudtrail_enabled(t)
}
