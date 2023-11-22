package compliance.cis_azure.rules.cis_7_2

import data.cis_azure.test_data
import data.compliance.policy.azure.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	eval_fail with input as test_data.generate_vm(null)
	eval_fail with input as test_data.generate_vm({})
	eval_fail with input as test_data.generate_vm({"data": "in", "unknown": "format"})
	eval_fail with input as test_data.generate_vm({"id": ""})
}

test_pass if {
	eval_pass with input as test_data.generate_vm(test_data.valid_managed_disk)
}

test_not_evaluated if {
	not_eval with input as {}
	not_eval with input as {"subType": "other-type", "resource": {"properties": {"storageProfile": {"osDisk": {"managedDisk": {"id": "some-id"}}}}}}
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
