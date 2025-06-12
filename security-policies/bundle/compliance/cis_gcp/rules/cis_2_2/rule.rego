package compliance.cis_gcp.rules.cis_2_2

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import future.keywords.if
import future.keywords.in

# Ensure That Sinks Are Configured for All Log Entries.
finding := result if {
	data_adapter.is_logging_asset

	passed := [r | r := input.resource.log_sinks[_]; is_sink_without_filter(r)]
	failed := [r | r := input.resource.log_sinks[_]; not is_sink_without_filter(r)]
	ok := count(passed) > 0

	evidence := common.get_evidence(ok, passed, failed)

	result := common.generate_result_without_expected(
		common.calculate_result(ok),
		{"log_sinks": {json.filter(c, ["name", "resource/data/filter"]) | c := evidence[_]}},
	)
}

# We recieved all sinks configured at the project, folder and org level.
# We check if any of the sinks are configured without a filter.
is_sink_without_filter(sink) if {
	not sink.resource.data.filter
} else := false
