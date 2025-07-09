package compliance.cis_gcp.rules.cis_4_7

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import future.keywords.if

# Ensure VM Disks for Critical VMs Are Encrypted With Customer-Supplied Encryption Keys (CSEK)
finding := result if {
	# filter
	data_adapter.is_compute_disk

	# set result
	result := common.generate_evaluation_result(common.calculate_result(is_disk_encrypted_with_csek))
}

is_disk_encrypted_with_csek if {
	data_adapter.resource.data.diskEncryptionKey.sha256 != ""
} else := false
