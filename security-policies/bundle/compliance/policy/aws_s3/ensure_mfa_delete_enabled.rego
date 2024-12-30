package compliance.policy.aws_s3.ensure_mfa_delete_enabled

import data.compliance.lib.common as lib_common
import data.compliance.policy.aws_s3.data_adapter
import future.keywords.if

default rule_evaluation := false

rule_evaluation if {
	bucket_versioning := data_adapter.bucket_versioning
	bucket_versioning.Enabled == true
	bucket_versioning.MfaDelete == true
}

finding := result if {
	data_adapter.is_s3
	not data_adapter.bucket_versioning == null

	result := lib_common.generate_result_without_expected(
		lib_common.calculate_result(rule_evaluation),
		{"BucketVersioning": data_adapter.bucket_versioning},
	)
}
