package compliance.cis_aws.rules.cis_1_15

import data.compliance.lib.common
import data.compliance.policy.aws_iam.data_adapter
import future.keywords.if

# Ensure IAM Users Receive Permissions Only Through Groups
finding := result if {
	# filter
	data_adapter.is_iam_user

	# set result
	user := data_adapter.iam_user
	result := common.generate_result_without_expected(
		common.calculate_result((count(user.attached_policies) + count(user.inline_policies)) == 0),
		{"IAM User:": user},
	)
}
