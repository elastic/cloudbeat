package compliance.cis_aws.rules.cis_2_3_2

import data.compliance.lib.common as lib_common
import data.compliance.policy.aws_rds.data_adapter
import future.keywords.if

finding := result if {
	data_adapter.is_rds

	result := lib_common.generate_result_without_expected(
		lib_common.calculate_result(data_adapter.auto_minor_version_upgrade == true),
		{"AutoMinorVersionUpgrade": data_adapter.auto_minor_version_upgrade},
	)
}
