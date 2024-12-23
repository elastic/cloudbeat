package compliance.cis_aws.rules.cis_3_9

import data.compliance.lib.common
import data.compliance.policy.aws_ec2.data_adapter
import future.keywords.if

default rule_evaluation := false

# Ensure VPC flow logging is enabled in all VPCs.
finding := result if {
	# filter
	data_adapter.is_vpc_policy

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(rule_evaluation),
		input.resource,
	)
}

rule_evaluation if {
	count(input.resource.flow_logs) > 0
}
