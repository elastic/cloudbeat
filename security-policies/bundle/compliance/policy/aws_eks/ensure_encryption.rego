package compliance.policy.aws_eks.ensure_encryption

import data.compliance.lib.common
import data.compliance.policy.aws_eks.data_adapter
import future.keywords.if

# Verify that there is a non empty encryption configuration
is_encrypted(cluster) if {
	cluster.EncryptionConfig
	count(cluster.EncryptionConfig) > 0
} else := false

# Ensure there Kuberenetes secrets are encrypted
finding := result if {
	# filter
	data_adapter.is_aws_eks

	rule_evaluation := is_encrypted(data_adapter.cluster)

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(rule_evaluation),
		{"encryption_config": data_adapter.cluster},
	)
}
