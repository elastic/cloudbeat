package compliance.cis_azure.rules.cis_1_23

import data.compliance.policy.azure.data_adapter
import data.lib.test
import data.cis_azure.test_data

generate_role_defenition_with_properties(properties) = {
	"properties": properties,
	"type": "microsoft.authorization/roledefinitions",
}

generate_role_defenitions(assets) = {
	"subType": "azure-role-definition",
	"resource": assets,
}

test_violation {
	eval_fail with input as generate_role_defenitions([generate_role_defenition_with_properties({"type": "CustomRole"})])
	eval_fail with input as generate_role_defenitions([
		generate_role_defenition_with_properties({"type": "CustomRole"}),
		generate_role_defenition_with_properties({"type": "BuiltInRole"}),
	])
}

test_pass {
	eval_pass with input as generate_role_defenitions([])
	eval_pass with input as generate_role_defenitions([generate_role_defenition_with_properties({"type": "BuiltInRole"})])
}

test_not_evaluated {
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
