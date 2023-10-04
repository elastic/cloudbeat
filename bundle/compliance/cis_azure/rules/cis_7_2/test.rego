package compliance.cis_azure.rules.cis_7_2

import data.cis_azure.test_data
import data.compliance.policy.azure.data_adapter
import data.lib.test

test_violation {
	eval_fail with input as test_data.generate_vm(null)
	eval_fail with input as test_data.generate_vm({})
	eval_fail with input as test_data.generate_vm({"data": "in", "unknown": "format"})
	eval_fail with input as test_data.generate_vm({"id": ""})
}

test_pass {
	eval_pass with input as test_data.generate_vm(test_data.valid_managed_disk)
}

test_not_evaluated {
	not_eval with input as {}
	not_eval with input as {"subType": "other-type", "resource": {"properties": {"storageProfile": {"osDisk": {"managedDisk": {"id": "some-id"}}}}}}
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
