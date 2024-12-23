package compliance.cis_aws.rules.cis_2_3_1

import data.compliance.lib.common as lib_common
import data.compliance.policy.aws_rds.data_adapter
import future.keywords.if

finding := result if {
	data_adapter.is_rds

	result := lib_common.generate_result_without_expected(
		lib_common.calculate_result(data_adapter.storage_encrypted == true),
		{"StorageEncrypted": data_adapter.storage_encrypted},
	)
}
