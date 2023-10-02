package compliance.cis_azure.rules.cis_7_4

import data.compliance.policy.azure.data_adapter
import data.lib.test

generate_encryption_settings(type) = {
	"diskEncryptionSetId": "/subscriptions/dead-beef/resourceGroups/RESOURCEGROUP/providers/Microsoft.Compute/diskEncryptionSets/double-disk-encryption-set",
	"type": type,
}

generate_unattached_disk_with_encryption(settings) = generate_disk_with_encryption("Unattached", settings)

generate_disk_with_encryption(state, settings) = {
	"type": "azure-disk",
	"subType": "",
	"resource": {
		"id": "/subscriptions/dead-beef/resourceGroups/resourceGroup/providers/Microsoft.Compute/disks/unattached-disk",
		"location": "eastus",
		"name": "unattached-disk",
		"properties": {
			"creationData": {"createOption": "Empty"},
			"dataAccessAuthMode": "None",
			"diskIOPSReadWrite": 500,
			"diskMBpsReadWrite": 60,
			"diskSizeBytes": 4294967296,
			"diskSizeGB": 4,
			"diskState": state,
			"encryption": settings,
			"networkAccessPolicy": "DenyAll",
			"provisioningState": "Succeeded",
			"publicNetworkAccess": "Disabled",
			"timeCreated": "2023-09-28T19:05:41.631Z",
			"uniqueId": "12345-abcdef",
		},
		"resource_group": "resourceGroup",
		"subscription_id": "dead-beef",
		"tenant_id": "beef-dead",
		"type": "microsoft.compute/disks",
	},
}

test_violation {
	eval_fail with input as generate_unattached_disk_with_encryption(null)
	eval_fail with input as generate_unattached_disk_with_encryption({"data": "in", "unknown": "format"})
	eval_fail with input as generate_unattached_disk_with_encryption(generate_encryption_settings("EncryptionAtRestWithPlatformKey"))
	eval_fail with input as generate_unattached_disk_with_encryption(generate_encryption_settings("InvalidValue"))
}

test_pass {
	eval_pass with input as generate_unattached_disk_with_encryption(generate_encryption_settings("EncryptionAtRestWithCustomerKey"))
	eval_pass with input as generate_unattached_disk_with_encryption(generate_encryption_settings("EncryptionAtRestWithPlatformAndCustomerKeys"))
}

test_not_evaluated {
	not_eval with input as {}
	not_eval with input as {"type": "other-type", "resource": {"encryption": {}}}
	not_eval with input as generate_disk_with_encryption("Attached", generate_encryption_settings("EncryptionAtRestWithPlatformAndCustomerKeys"))
	not_eval with input as generate_disk_with_encryption("OtherState", generate_encryption_settings("EncryptionAtRestWithPlatformAndCustomerKeys"))
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
