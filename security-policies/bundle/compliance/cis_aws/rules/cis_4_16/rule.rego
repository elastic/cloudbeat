package compliance.cis_aws.rules.cis_4_16

import data.compliance.lib.common
import data.compliance.policy.aws_securityhub.data_adapter
import future.keywords.if

default rule_evaluation := false

finding := result if {
	# filter
	data_adapter.is_securityhub_subType

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(rule_evaluation),
		input.resource,
	)
}

rule_evaluation := data_adapter.securityhub_resource.Enabled
