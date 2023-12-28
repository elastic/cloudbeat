package compliance.cis_gcp.rules.cis_1_5

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter

import data.compliance.policy.gcp.iam.ensure_not_admin_roles as audit

import future.keywords.if

# service_accounts filters out members that are not service accounts
# GCP has 2 types of SA: user managed and google managed
# differinitating between the two based on the member email is not possible
# and this CIS rule only applies to user managed service accounts
# so a google managed SA may produce a failed finding, even though it is not applicable
service_accounts := [{"members": filtered_members, "role": v.role} |
	v := data_adapter.iam_policy.bindings[_]
	filtered_members := [m |
		m = v.members[_]
		startswith(m, "serviceAccount:")
	]
	count(filtered_members) > 0
]

finding = result if {
	data_adapter.is_cloud_resource_manager_project
	data_adapter.has_policy
	count(service_accounts) > 0

	result := common.generate_result_without_expected(
		common.calculate_result(audit.is_not_admin_roles(service_accounts)),
		evidence,
	)
}

# maps the service accounts array to an object with keys as roles and values as members
# this makes it easier to see which service accounts has which role
evidence = admin_roles if {
	admin_roles := {role: members |
		entry := service_accounts[_]
		role := entry.role
		members := entry.members
		regex.match(`(.*Admin|.*admin|roles/(editor|owner))`, role)
	}
	count(admin_roles) > 0
} else = {role: members |
	entry := service_accounts[_]
	role := entry.role
	members := entry.members
}
