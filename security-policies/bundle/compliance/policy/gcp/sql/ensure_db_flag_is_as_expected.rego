package compliance.policy.gcp.sql.ensure_db_flag

import data.compliance.policy.gcp.data_adapter
import future.keywords.every
import future.keywords.if
import future.keywords.in

is_flag_configured_as_expected(flag_name, expected_vals) if {
	some db_flag in data_adapter.resource.data.settings.databaseFlags
	db_flag.name == flag_name

	# not all expected values needs to be present, one is sufficient
	some expected_val in expected_vals
	db_flag.value == expected_val
} else := false

is_flag_exists(flag_name) if {
	some db_flag in data_adapter.resource.data.settings.databaseFlags
	db_flag.name == flag_name
} else := false

is_flag_limited(flag_name) if {
	some db_flag in data_adapter.resource.data.settings.databaseFlags
	db_flag.name == flag_name
	db_flag.value != 0
} else := false
