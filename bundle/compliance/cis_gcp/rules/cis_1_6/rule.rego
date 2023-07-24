package compliance.cis_gcp.rules.cis_1_6

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import data.compliance.policy.gcp.iam.ensure_role_not_service_account_user as audit
import future.keywords.if

finding = result if {
	data_adapter.is_iam_service_account
	data_adapter.has_policy

	result := common.generate_result_without_expected(
		common.calculate_result(audit.is_role_not_service_account_user),
		data_adapter.iam_policy.bindings[i].role,
	)
}
