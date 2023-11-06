package compliance.cis_gcp.rules.cis_4_7

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter

# Ensure VM Disks for Critical VMs Are Encrypted With Customer-Supplied Encryption Keys (CSEK)
finding = result {
	# filter
	data_adapter.is_compute_disk

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_disk_encrypted_with_csek),
		{"Compute instance": data_adapter.resource},
	)
}

is_disk_encrypted_with_csek {
	data_adapter.resource.data.diskEncryptionKey.sha256 != ""
} else = false
