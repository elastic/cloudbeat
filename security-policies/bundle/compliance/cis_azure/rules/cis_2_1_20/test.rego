package compliance.cis_azure.rules.cis_2_1_20

import data.cis_azure.test_data
import data.compliance.policy.azure.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	eval_fail with input as test_data.generate_security_contacts([test_data.generate_single_security_contact(
		"default",
		prop_notification_sources([{
			"minimalSeverity": "Medium",
			"sourceType": "Another",
		}]),
	)])

	eval_fail with input as test_data.generate_security_contacts([test_data.generate_single_security_contact(
		"default",
		prop_notification_sources([{"sourceType": "Alert"}]),
	)])

	eval_fail with input as test_data.generate_security_contacts([
		test_data.generate_single_security_contact(
			"non-default",
			prop_notification_sources([{
				"minimalSeverity": "High",
				"sourceType": "Alert",
			}]),
		),
		test_data.generate_single_security_contact(
			"default",
			prop_notification_sources([{
				"minimalSeverity": "Wrong value",
				"sourceType": "Alert",
			}]),
		),
	])
}

test_pass if {
	eval_pass with input as test_data.generate_security_contacts([test_data.generate_single_security_contact(
		"default",
		prop_notification_sources([{
			"minimalSeverity": "High",
			"sourceType": "Alert",
		}]),
	)])

	eval_pass with input as test_data.generate_security_contacts([test_data.generate_single_security_contact(
		"default",
		prop_notification_sources([{
			"minimalSeverity": "Medium",
			"sourceType": "Alert",
		}]),
	)])

	eval_pass with input as test_data.generate_security_contacts([test_data.generate_single_security_contact(
		"default",
		prop_notification_sources([{
			"minimalSeverity": "Low",
			"sourceType": "Alert",
		}]),
	)])

	eval_pass with input as test_data.generate_security_contacts([
		test_data.generate_single_security_contact(
			"non-default",
			prop_notification_sources([{
				"minimalSeverity": "Low",
				"sourceType": "Alert",
			}]),
		),
		test_data.generate_single_security_contact(
			"default",
			prop_notification_sources([{
				"minimalSeverity": "High",
				"sourceType": "Alert",
			}]),
		),
	])
}

test_not_evaluated if {
	not_eval with input as test_data.not_eval_non_exist_type
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

prop_notification_sources(notificationsSource) := {"notificationsSources": notificationsSource}
