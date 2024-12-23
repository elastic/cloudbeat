package compliance.policy.azure.storage_account.ensure_service_log

import future.keywords.every
import future.keywords.if
import future.keywords.in

service_diagnostic_settings_log_rwd_enabled(serviceDiagnosticSettings) if {
	# Ensure all categories exist and are enabled
	every category in ["StorageRead", "StorageWrite", "StorageDelete"] {
		some log in serviceDiagnosticSettings.properties.logs
		log.enabled == true
		log.category = category
	}
} else := false
