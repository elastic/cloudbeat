package compliance.cis_aws.rules.cis_1_4

import data.compliance.lib.common
import data.compliance.policy.aws_iam.data_adapter
import future.keywords.if

# Ensure no 'root' user account access key exists.
finding := result if {
	# filter
	data_adapter.is_root_user

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(count(data_adapter.active_access_keys) == 0),
		{"IAM User:": data_adapter.iam_user},
	)
}
