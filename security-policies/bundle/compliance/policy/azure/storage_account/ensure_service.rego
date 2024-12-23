package compliance.policy.azure.storage_account.ensure_service

import data.compliance.policy.azure.data_adapter
import future.keywords.if
import future.keywords.in

is_service_included(service) if {
	data_adapter.network_acls.defaultAction == "Deny"
	data_adapter.network_acls.bypass == service
} else := false

evaluate_service(service) := r if {
	data_adapter.network_acls
	r = is_service_included(service)
}
