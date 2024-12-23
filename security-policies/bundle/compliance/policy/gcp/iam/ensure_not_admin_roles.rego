package compliance.policy.gcp.iam.ensure_not_admin_roles

import future.keywords.if
import future.keywords.in

# role of members group associated with each serviceaccount does not contain
# *Admin or *admin or does not match roles/editor or does not match roles/owner.
is_not_admin_roles(service_accounts) if {
	admin_roles := {v | v := service_accounts[_].role; regex.match(`(.*Admin|.*admin|roles/(editor|owner))`, v)}
	count(admin_roles) == 0
} else := false
