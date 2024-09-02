package compliance.cis_azure.rules.cis_2_1_20

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.if
import future.keywords.in

finding = result if {
	# filter
	data_adapter.is_security_contacts

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(notification_alert_high),
		{"Resource": data_adapter.resource},
	)
}

default notification_alert_high = false

notification_alert_high if {
	# Ensure at least one Security Contact Settings exists and alertNotifications severity is set to high, low, or medium.
	some security_contact in data_adapter.resource

	security_contact.name == "default"

	some notification_source in security_contact.properties.notificationsSources

	lower(notification_source.sourceType) == "alert"
	lower(notification_source.minimalSeverity) in ["low", "medium", "high"]
}
