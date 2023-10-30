package compliance.cis_gcp.rules.cis_1_5

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import data.compliance.policy.gcp.iam.ensure_user_not_editor_or_owner as audit
import future.keywords.every
import future.keywords.if

finding = result if {
	data_adapter.is_iam_service_account
	data_adapter.has_policy

	result := common.generate_result_without_expected(
		common.calculate_result(audit.is_user_owner_or_editor),
		data_adapter.iam_policy,
	)
}
