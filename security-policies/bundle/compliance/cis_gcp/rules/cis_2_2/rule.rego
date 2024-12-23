package compliance.cis_gcp.rules.cis_2_2

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import future.keywords.if
import future.keywords.in

# Ensure That Sinks Are Configured for All Log Entries.
finding := result if {
	data_adapter.is_logging_asset

	result := common.generate_result_without_expected(
		common.calculate_result(is_sink_without_filter),
		input.resource.log_sinks,
	)
}

# We recieved all sinks configured at the project, folder and org level.
# We check if any of the sinks are configured without a filter.
is_sink_without_filter if {
	some sink in input.resource.log_sinks
	not sink.resource.data.filter
} else := false
