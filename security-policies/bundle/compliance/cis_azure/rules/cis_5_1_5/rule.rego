package compliance.cis_azure.rules.cis_5_1_5

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_vault

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_vault_logging_enabled),
		{"Resource": data_adapter.resource},
	)
}

is_audit_category(i) if i.categoryGroup == "allLogs"

is_audit_category(i) if i.categoryGroup == "audit"

# AuditEvent category is in both categoryGroup "allLogs" and "audit"
is_audit_category(i) if i.category == "AuditEvent"

is_vault_logging_enabled if {
	entry = data_adapter.resource.extension.vaultDiagnosticSettings[i].properties
	entry.storageAccountId != null
	logs := entry.logs[i]
	logs.enabled == true
	is_audit_category(logs)
} else := false
