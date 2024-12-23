package compliance.cis_azure.rules.cis_5_1_2

import data.cis_azure.test_data
import data.compliance.policy.azure.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	# fail if no diagnostic settings exists
	eval_fail with input as test_data.generate_diagnostic_settings_empty

	# pass if not all categories are enabled
	eval_fail with input as test_data.generate_diagnostic_settings([component1])
	eval_fail with input as test_data.generate_diagnostic_settings([component1, component2])
}

test_pass if {
	# pass if all categories are selected to one component
	eval_pass with input as test_data.generate_diagnostic_settings([component3])

	# pass if all categories are selected aggregated to all diagnostic settings
	eval_pass with input as test_data.generate_diagnostic_settings([component4, component5])
}

test_not_evaluated if {
	# not_eval if the resource is not relevant
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

component1 := test_data.generate_diagnostic_setting_element(
	"sub1",
	"rcg1",
	"name1",
	test_data.generate_diagnostic_setting_element_logs({"Administrative": false, "Alert": false, "Policy": false, "Security": false}),
)

component2 := test_data.generate_diagnostic_setting_element(
	"sub1",
	"rcg1",
	"name2",
	test_data.generate_diagnostic_setting_element_logs({"Administrative": true, "Alert": false, "Policy": false, "Security": false}),
)

component3 := test_data.generate_diagnostic_setting_element(
	"sub1",
	"rcg1",
	"name3",
	test_data.generate_diagnostic_setting_element_logs({"Administrative": true, "Alert": true, "Policy": true, "Security": true}),
)

component4 := test_data.generate_diagnostic_setting_element(
	"sub1",
	"rcg1",
	"name3",
	test_data.generate_diagnostic_setting_element_logs({"Administrative": true, "Alert": true, "Policy": false, "Security": false}),
)

component5 := test_data.generate_diagnostic_setting_element(
	"sub1",
	"rcg1",
	"name3",
	test_data.generate_diagnostic_setting_element_logs({"Administrative": false, "Alert": false, "Policy": true, "Security": true}),
)
