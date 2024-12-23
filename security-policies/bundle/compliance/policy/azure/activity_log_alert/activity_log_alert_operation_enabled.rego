package compliance.policy.azure.activity_log_alert.activity_log_alert_operation_enabled

import data.compliance.policy.azure.data_adapter
import future.keywords.if
import future.keywords.in

# Ensure that the Activity Log Alert for specific Administrative operations exists, is enabled, and meets specific conditions
activity_log_alert_operation_enabled(operation_names, categories) if {
	# Ensure at least one Activity Log Alert passes the following conditions
	some activity_log_alert in data_adapter.activity_log_alerts

	# Ensure the Activity Log Alert is enabled
	activity_log_alert.properties.enabled

	# Ensure the operationName condition exists for specific operations
	operation_name_exists := [condition |
		condition := activity_log_alert.properties.condition.allOf[_]
		condition.field == "operationName"
		condition.equals == operation_names[_]
	]
	count(operation_name_exists) >= 1

	# Ensure the category condition exists and is set to the allowed categories
	category_exists := [condition |
		condition := activity_log_alert.properties.condition.allOf[_]
		condition.field == "category"
		condition.equals == categories[_]
	]
	count(category_exists) == 1

	# Ensure there is an action group assigned (Notification to the appropriate personnel)
	activity_log_alert.properties.actions != null
} else := false
