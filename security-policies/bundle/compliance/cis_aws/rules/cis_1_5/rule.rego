package compliance.cis_aws.rules.cis_1_5

import data.compliance.lib.common
import data.compliance.policy.aws_iam.data_adapter
import future.keywords.if

# Ensure MFA is enabled for the 'root' user account.
finding := result if {
	# filter
	data_adapter.is_root_user

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(data_adapter.iam_user.mfa_active),
		{"IAM User:": data_adapter.iam_user},
	)
}
