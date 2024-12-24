package compliance.cis_aws.rules.cis_1_12

import data.compliance.lib.common
import data.compliance.policy.aws_iam.data_adapter
import data.compliance.policy.aws_iam.validate_credentials as audit
import future.keywords.if

# Ensure credentials unused for 45 days or greater are disabled
finding := result if {
	# filter
	data_adapter.is_iam_user

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(audit.validate_credentials),
		{"IAM User:": data_adapter.iam_user},
	)
}
