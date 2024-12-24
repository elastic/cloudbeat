package compliance.policy.azure.storage_account.ensure_tls_version

import data.compliance.policy.azure.data_adapter
import future.keywords.if

is_tls_version(version) if {
	data_adapter.properties.minimumTlsVersion == version
} else := false

is_tls_configured(version) := r if {
	data_adapter.properties.minimumTlsVersion
	r = is_tls_version(version)
}
