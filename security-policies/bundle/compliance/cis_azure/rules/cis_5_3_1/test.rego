package compliance.cis_azure.rules.cis_5_3_1

import data.cis_azure.test_data
import data.compliance.policy.azure.data_adapter
import data.lib.test

test_violation {
	# fail if no insights component exists
	eval_fail with input as test_data.generate_insights_components_empty
}

test_pass {
	# pass if one insights component exists
	eval_pass with input as test_data.generate_insights_components([component1])

	# pass if more than one insights component exist
	eval_pass with input as test_data.generate_insights_components([component1, component2])
}

test_not_evaluated {
	# not_eval if the resource is not relevant
	not_eval with input as test_data.not_eval_resource
}

eval_fail {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval {
	not finding with data.benchmark_data_adapter as data_adapter
}

component1 = test_data.generate_insights_component("rcg1", "cmp1")

component2 = test_data.generate_insights_component("rcg1", "cmp1")
