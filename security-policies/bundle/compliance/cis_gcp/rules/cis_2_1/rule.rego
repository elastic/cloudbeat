package compliance.cis_gcp.rules.cis_2_1

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import future.keywords.if
import future.keywords.in

finding := result if {
	data_adapter.is_policies_resource

	passed := [r | r := input.resource[_]; is_cloud_logging_configured(r)]
	failed := [r | r := input.resource[_]; not is_cloud_logging_configured(r)]
	ok := count(passed) > 0

	result := common.generate_result_without_expected(
		common.calculate_result(ok),
		{"policies": {json.filter(c, ["name", "iam_policy/audit_configs"]) | c := common.get_evidence(ok, passed, failed)[_]}},
	)
}

is_cloud_logging_configured(resource) if {
	policy := resource.iam_policy
	has_read_write_logs(policy)
	not has_exempted_members(policy)
} else := false

has_read_write_logs(policy) if {
	log_types := {t | t = policy.audit_configs[i].audit_log_configs[j].log_type}
	1 in log_types # "ADMIN_READ"
	2 in log_types # "DATA_WRITE"
	3 in log_types # "DATA_READ"
	policy.audit_configs[_].service == "allServices"
} else := false

has_exempted_members(policy) if {
	configs := policy.audit_configs[_].audit_log_configs[_]
	count(configs.exempted_members) > 0
} else := false
