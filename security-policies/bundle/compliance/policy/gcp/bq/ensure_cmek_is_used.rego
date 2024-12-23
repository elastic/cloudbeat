package compliance.policy.gcp.bq.ensure_cmek_is_used

import data.compliance.policy.gcp.data_adapter
import future.keywords.if

default is_cmek_used := false

is_cmek_used if {
	data_adapter.resource.data.defaultEncryptionConfiguration.kmsKeyName
}
