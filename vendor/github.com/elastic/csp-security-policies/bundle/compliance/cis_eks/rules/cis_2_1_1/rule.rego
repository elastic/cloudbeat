package compliance.cis_eks.rules.cis_2_1_1

import data.compliance.cis_eks.data_adapter
import data.compliance.lib.assert
import data.compliance.lib.common

# Ensure that all audit logs are enabled
finding = result {
	# filter
	data_adapter.is_aws_eks

	# evaluate
	cluster_logging := input.resource.Cluster.Logging.ClusterLogging
	disabled_logs := [log | assert.is_false(cluster_logging[index].Enabled); log = cluster_logging[index].Types[_]]
	rule_evaluation := count(disabled_logs) == 0

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"disabled_logs": disabled_logs},
	}
}
