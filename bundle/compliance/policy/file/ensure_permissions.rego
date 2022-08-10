package compliance.policy.file.ensure_permissions

import data.compliance.lib.common as lib_common
import data.compliance.policy.file.common as file_common
import data.compliance.policy.file.data_adapter

finding(user, group, other) := result {
	mode := data_adapter.filemode
	rule_evaluation := file_common.file_permission_match(mode, user, group, other)

	# set result
	result := lib_common.generate_result(
		lib_common.calculate_result(rule_evaluation),
		{"filemode": mode},
		{"filemode": ((user * 100) + (group * 10)) + other},
	)
}

path_filter(name) := file_common.file_in_path(name, data_adapter.file_path)

filename_filter(name) := data_adapter.filename == name

filename_suffix_filter(suffix) := endswith(data_adapter.filename, suffix)
