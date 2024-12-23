package compliance.policy.aws_eks.ensure_logs_enabled

import data.compliance.lib.assert
import data.compliance.lib.common
import data.compliance.policy.aws_eks.data_adapter
import future.keywords.if

# Ensure that all audit logs are enabled
finding := result if {
	# filter
	data_adapter.is_aws_eks

	# evaluate
	cluster_logging := data_adapter.cluster.Logging.ClusterLogging
	disabled_logs := [log | assert.is_false(cluster_logging[index].Enabled); log = cluster_logging[index].Types[_]]
	rule_evaluation := count(disabled_logs) == 0

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(rule_evaluation),
		{"disabled_logs": disabled_logs},
	)
}
