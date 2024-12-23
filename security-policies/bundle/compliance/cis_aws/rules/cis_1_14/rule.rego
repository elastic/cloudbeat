package compliance.cis_aws.rules.cis_1_14

import data.compliance.lib.common
import data.compliance.policy.aws_iam.data_adapter
import data.compliance.policy.aws_iam.verify_keys_rotation as audit
import future.keywords.if

# Ensure access keys are rotated every 90 days or less
finding := result if {
	# filter
	data_adapter.is_iam_user

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(audit.verify_rotation),
		{"IAM User:": data_adapter.iam_user},
	)
}
