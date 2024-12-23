package compliance.policy.gcp.compute.ensure_default_sa

import data.compliance.policy.gcp.data_adapter
import future.keywords.every
import future.keywords.if

is_default_sa(sa) if {
	endswith(sa.email, "-compute@developer.gserviceaccount.com")
}

is_default_sa_with_access(sa) if {
	is_default_sa(sa)
	some scope in sa.scopes
	scope == "https://www.googleapis.com/auth/cloud-platform"
}

sa_is_default if {
	not data_adapter.is_gke_instance(data_adapter.resource.data)
	some sa in data_adapter.resource.data.serviceAccounts
	is_default_sa(sa)
} else := false

sa_is_default_with_full_access if {
	not data_adapter.is_gke_instance(data_adapter.resource.data)
	some sa in data_adapter.resource.data.serviceAccounts
	is_default_sa_with_access(sa)
} else := false
