package compliance.policy.file.ensure_permissions

import data.compliance.lib.common as lib_common
import data.compliance.policy.file.common as file_common
import data.compliance.policy.file.data_adapter

finding(rule_evaluation) := result {
	# set result
	result := lib_common.generate_result(
		lib_common.calculate_result(rule_evaluation.evaluation),
		{"filemode": rule_evaluation.mode},
		{"filemode": sprintf("%d%d%d", [rule_evaluation.user, rule_evaluation.group, rule_evaluation.other])},
	)
}

path_filter(name) := file_common.file_in_path(name, data_adapter.file_path)

filename_filter(name) := data_adapter.filename == name

filename_suffix_filter(suffix) := endswith(data_adapter.filename, suffix)

file_permission_match(user, group, other) := result {
	mode := data_adapter.filemode
	result := {
		"evaluation": file_common.file_permission_match(mode, user, group, other),
		"mode": mode,
		"user": user,
		"group": group,
		"other": other,
	}
}

file_permission_match_exact(user, group, other) := result {
	mode := data_adapter.filemode
	result := {
		"evaluation": file_common.file_permission_match_exact(mode, user, group, other),
		"mode": mode,
		"user": user,
		"group": group,
		"other": other,
	}
}
