package compliance.cis_azure.rules.cis_5_2_1

import data.cis_azure.test_data
import data.compliance.policy.azure.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	# fail if no alert exists
	eval_fail with input as test_data.generate_activity_log_alerts_no_alerts

	# fail if no alert matches the rule
	eval_fail with input as test_data.generate_activity_log_alerts([mismatch_alert])

	# fail if no alert matches the rule
	eval_fail with input as test_data.generate_activity_log_alerts([mismatch_alert_only_operation])

	# fail if no alert matches the rule
	eval_fail with input as test_data.generate_activity_log_alerts([mismatch_alert_only_category])

	# fail if no alert matches the rule
	eval_fail with input as test_data.generate_activity_log_alerts([mismatch_alert, mismatch_alert_only_operation, mismatch_alert_only_category])
}

test_pass if {
	# pass if the alert exists and is properly configured
	eval_pass with input as test_data.generate_activity_log_alerts([matching_alert])

	# pass if at least one alert exists and is properly configured
	eval_pass with input as test_data.generate_activity_log_alerts([matching_alert, mismatch_alert])
}

test_not_evaluated if {
	# not_eval if the resiurce is not relevant
	not_eval with input as test_data.not_eval_resource
}

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval if {
	not finding with data.benchmark_data_adapter as data_adapter
}

# test data
# alert rule that does not match the rule by operation and category
mismatch_alert := test_data.generate_activity_log_alert("mismatch_opreation", "mismatch_category")

# alert rule that does not match the rule by operation
mismatch_alert_only_operation := test_data.generate_activity_log_alert("mismatch_opreation", "Administrative")

# alert rule that does not match the rule by category
mismatch_alert_only_category := test_data.generate_activity_log_alert("Microsoft.Authorization/policyAssignments/write", "mismatch_category")

# alert rule that matches the rule
matching_alert := test_data.generate_activity_log_alert("Microsoft.Authorization/policyAssignments/write", "Administrative")
