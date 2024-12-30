package compliance.cis_aws.rules.cis_1_7

import data.compliance.lib.common
import data.compliance.policy.aws_iam.data_adapter
import data.compliance.policy.aws_iam.verify_user_usage as audit
import future.keywords.if

# Eliminate use of the 'root' user for administrative and daily tasks
# daily interpret as a day (24h)
finding := result if {
	# filter
	data_adapter.is_root_user

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(audit.verify_user_usage),
		{"IAM User:": data_adapter.iam_user},
	)
}
