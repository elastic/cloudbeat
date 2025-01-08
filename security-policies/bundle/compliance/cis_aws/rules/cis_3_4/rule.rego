package compliance.cis_aws.rules.cis_3_4

import data.compliance.lib.common
import data.compliance.policy.aws_cloudtrail.data_adapter
import data.compliance.policy.aws_cloudtrail.ensure_cloudwatch as audit
import future.keywords.if

# Ensure CloudTrail trails are integrated with CloudWatch Logs.
finding := result if {
	# filter
	data_adapter.is_single_trail

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(audit.ensure_cloudwatch_logs_enabled),
		input.resource,
	)
}
