package compliance.cis_azure.rules.cis_2_1_18

import data.cis_azure.test_data
import data.compliance.policy.azure.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	eval_fail with input as test_data.generate_security_contacts([test_data.generate_single_security_contact(
		"default",
		prop_notification_by_role({
			"state": "On",
			"roles": ["Admin"],
		}),
	)])

	eval_fail with input as test_data.generate_security_contacts([test_data.generate_single_security_contact(
		"default",
		prop_notification_by_role({
			"state": "Off",
			"roles": ["Owner"],
		}),
	)])

	eval_fail with input as test_data.generate_security_contacts([test_data.generate_single_security_contact(
		"default",
		prop_notification_by_role({"state": "On"}),
	)])

	eval_fail with input as test_data.generate_security_contacts([test_data.generate_single_security_contact(
		"non-default",
		prop_notification_by_role({
			"state": "On",
			"roles": ["Owner"],
		}),
	)])

	eval_fail with input as test_data.generate_security_contacts([
		test_data.generate_single_security_contact(
			"non-default",
			prop_notification_by_role({
				"state": "On",
				"roles": ["Owner"],
			}),
		),
		test_data.generate_single_security_contact(
			"default",
			prop_notification_by_role({
				"state": "On",
				"roles": ["Admin"],
			}),
		),
	])
}

test_pass if {
	eval_pass with input as test_data.generate_security_contacts([test_data.generate_single_security_contact(
		"default",
		prop_notification_by_role({
			"state": "On",
			"roles": ["Owner"],
		}),
	)])

	eval_pass with input as test_data.generate_security_contacts([test_data.generate_single_security_contact(
		"default",
		prop_notification_by_role({
			"state": "On",
			"roles": ["Owner", "Admin"],
		}),
	)])

	eval_pass with input as test_data.generate_security_contacts([
		test_data.generate_single_security_contact(
			"non-default",
			prop_notification_by_role({
				"state": "On",
				"roles": ["Admin"],
			}),
		),
		test_data.generate_single_security_contact(
			"default",
			prop_notification_by_role({
				"state": "On",
				"roles": ["Owner"],
			}),
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

prop_notification_by_role(notificationsByRole) := {"notificationsByRole": notificationsByRole}
