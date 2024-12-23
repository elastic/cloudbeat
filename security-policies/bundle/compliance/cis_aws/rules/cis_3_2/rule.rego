package compliance.cis_aws.rules.cis_3_2

import data.compliance.lib.common
import data.compliance.policy.aws_cloudtrail.data_adapter
import future.keywords.if

# Ensure CloudTrail log file validation is enabled.
finding := result if {
	# filter
	data_adapter.is_single_trail

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(data_adapter.trail.LogFileValidationEnabled),
		input.resource,
	)
}
