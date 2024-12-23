package compliance.cis_azure.rules.cis_5_3_1

import data.cis_azure.test_data
import data.compliance.policy.azure.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	# fail if no insights component exists
	eval_fail with input as test_data.generate_insights_components_empty
}

test_pass if {
	# pass if one insights component exists
	eval_pass with input as test_data.generate_insights_components([component1])

	# pass if more than one insights component exist
	eval_pass with input as test_data.generate_insights_components([component1, component2])
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

component1 := test_data.generate_insights_component("rcg1", "cmp1")

component2 := test_data.generate_insights_component("rcg2", "cmp2")
