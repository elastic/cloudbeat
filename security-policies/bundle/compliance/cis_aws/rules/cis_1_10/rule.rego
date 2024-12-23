package compliance.cis_aws.rules.cis_1_10

import data.compliance.lib.common
import data.compliance.policy.aws_iam.data_adapter
import data.compliance.policy.aws_iam.ensure_enabled_mfa as audit
import future.keywords.if

# Ensure multi-factor authentication (MFA) is enabled for all IAM users that have a console password.
finding := result if {
	# filter
	data_adapter.is_iam_user

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(audit.ensure_mfa_device),
		{"IAM User:": data_adapter.iam_user},
	)
}
