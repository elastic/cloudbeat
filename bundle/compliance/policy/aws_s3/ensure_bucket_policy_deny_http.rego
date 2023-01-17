package compliance.policy.aws_s3.ensure_bucket_policy_deny_http

import data.compliance.lib.common as lib_common
import data.compliance.policy.aws_s3.data_adapter

default rule_evaluation = false

rule_evaluation {
	statement := data_adapter.bucket_policy_statement
	statement.Condition.Bool["aws:SecureTransport"] == "false"
	statement.Action == "s3:*"
	statement.Effect == "Deny"
	statement.Principal == "*"
}

finding = result {
	data_adapter.is_s3

	result := lib_common.generate_result_without_expected(
		lib_common.calculate_result(rule_evaluation),
		{"BucketPolicy": data_adapter.bucket_policy},
	)
}
