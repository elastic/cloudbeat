package compliance.policy.gcp.iam.ensure_role_not_service_account_user

import data.compliance.policy.gcp.data_adapter
import future.keywords.if

default is_role_not_service_account_user := false

is_role_not_service_account_user if {
	role := data_adapter.iam_policy.bindings[i].role
	role != "roles/iam.serviceAccountUser"
	role != "roles/iam.serviceAccountTokenCreator"
}
