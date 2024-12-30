package compliance.policy.aws_s3.ensure_bucket_policy_deny_http

import data.compliance.lib.common as lib_common
import data.compliance.policy.aws_s3.data_adapter
import future.keywords.if
import future.keywords.in

default rule_evaluation := false

rule_evaluation if {
	some statement in data_adapter.bucket_policy_statements
	statement.Condition.Bool["aws:SecureTransport"] == "false"
	statement.Action == "s3:*"
	statement.Effect == "Deny"
	statement.Principal == "*"
}

finding := result if {
	data_adapter.is_s3
	not data_adapter.bucket_policy == null

	result := lib_common.generate_result_without_expected(
		lib_common.calculate_result(rule_evaluation),
		{"BucketPolicy": data_adapter.bucket_policy},
	)
}
