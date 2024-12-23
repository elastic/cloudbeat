package compliance.policy.file.ensure_ownership

import data.compliance.lib.common as lib_common
import data.compliance.policy.file.common as file_common
import data.compliance.policy.file.data_adapter
import future.keywords.if

finding(owner_user, owner_group) := result if {
	user = data_adapter.owner_user
	group = data_adapter.owner_group
	rule_evaluation := file_common.file_ownership_match(user, group, owner_user, owner_group)

	# set result
	result := lib_common.generate_result(
		lib_common.calculate_result(rule_evaluation),
		{"owner": user, "group": group},
		{"owner": owner_user, "group": owner_group},
	)
}

path_filter(name) := file_common.file_in_path(name, data_adapter.file_path)

filename_filter(name) := data_adapter.filename == name
