package compliance.cis_gcp.rules.cis_1_17

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import future.keywords.if

default has_cusomter_encrypted_key := false

finding := result if {
	data_adapter.is_dataproc_cluster

	result := common.generate_evaluation_result(common.calculate_result(has_cusomter_encrypted_key))
}

has_cusomter_encrypted_key if {
	data_adapter.resource.data.config.encryptionConfig.gcePdKmsKeyName != null
	data_adapter.resource.data.config.encryptionConfig.gcePdKmsKeyName != ""
}
