package compliance.cis_eks.rules.cis_5_3_1

import data.compliance.cis_eks.data_adapter
import data.compliance.lib.common

default rule_evaluation = false

# Verify that there is a non empty encryption configuration
rule_evaluation {
	input.resource.Cluster.EncryptionConfig
	count(input.resource.Cluster.EncryptionConfig) > 0
}

# Ensure there Kuberenetes secrets are encrypted
finding = result {
	# filter
	data_adapter.is_aws_eks

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"encryption_config": input.resource.Cluster},
	}
}
