package compliance.cis_gcp.rules.cis_1_11

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import data.compliance.policy.gcp.iam.ensure_admin_without_multiple_roles as audit
import future.keywords.if

default admin_has_other_role := false

finding := result if {
	data_adapter.is_cloud_resource_manager_project
	data_adapter.has_policy

	result := common.generate_result_without_expected(
		# admin_has_other_role is an aggregation of the three checks above
		# if an admin user has any of those other roles, the rule fails
		common.calculate_result(admin_has_other_role == false),
		data_adapter.iam_policy,
	)
}

# check admin is not also cryptoKeyEncrypter
admin_has_other_role if {
	audit.admin_has_multiple_roles("roles/cloudkms.admin", "roles/cloudkms.cryptoKeyEncrypter")
}

# check admin is not also cryptoKeyDecrypter
admin_has_other_role if {
	audit.admin_has_multiple_roles("roles/cloudkms.admin", "roles/cloudkms.cryptoKeyDecrypter")
}

# check admin is not also cryptoKeyEncrypterDecrypter
admin_has_other_role if {
	audit.admin_has_multiple_roles("roles/cloudkms.admin", "roles/cloudkms.cryptoKeyEncrypterDecrypter")
}
