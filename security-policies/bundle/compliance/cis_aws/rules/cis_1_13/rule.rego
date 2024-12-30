package compliance.cis_aws.rules.cis_1_13

import data.compliance.lib.common
import data.compliance.policy.aws_iam.data_adapter
import future.keywords.if

# Ensure that there is only a single active access key per user.
finding := result if {
	# filter
	data_adapter.is_iam_user

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(count(data_adapter.active_access_keys) < 2),
		{"IAM User:": data_adapter.iam_user},
	)
}
