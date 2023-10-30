package compliance.cis_azure.rules.cis_7_4

import data.cis_azure.test_data
import data.compliance.policy.azure.data_adapter
import data.lib.test

test_violation {
	eval_fail with input as test_data.generate_unattached_disk_with_encryption({})
	eval_fail with input as test_data.generate_unattached_disk_with_encryption({"encryption": {}})
	eval_fail with input as test_data.generate_unattached_disk_with_encryption({"data": "in", "unknown": "format"})
	eval_fail with input as test_data.generate_unattached_disk_with_encryption(test_data.generate_disk_encryption_settings("EncryptionAtRestWithPlatformKey"))
	eval_fail with input as test_data.generate_unattached_disk_with_encryption(test_data.generate_disk_encryption_settings("InvalidValue"))
}

test_pass {
	eval_pass with input as test_data.generate_unattached_disk_with_encryption(test_data.generate_disk_encryption_settings("EncryptionAtRestWithCustomerKey"))
	eval_pass with input as test_data.generate_unattached_disk_with_encryption(test_data.generate_disk_encryption_settings("EncryptionAtRestWithPlatformAndCustomerKeys"))
}

test_not_evaluated {
	not_eval with input as {}
	not_eval with input as test_data.not_eval_resource
	not_eval with input as {"subType": "other-type", "resource": {"encryption": {}}}
	not_eval with input as test_data.generate_attached_disk_with_encryption(test_data.generate_disk_encryption_settings("EncryptionAtRestWithPlatformAndCustomerKeys"))
	not_eval with input as test_data.generate_disk_with_encryption("OtherState", test_data.generate_disk_encryption_settings("EncryptionAtRestWithPlatformAndCustomerKeys"))
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
