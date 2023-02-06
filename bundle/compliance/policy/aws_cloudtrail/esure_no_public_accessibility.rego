package compliance.policy.aws_cloudtrail.no_public_bucket_access

import data.compliance.lib.common
import data.compliance.policy.aws_cloudtrail.data_adapter
import future.keywords.every

default no_public_log_access = false

no_public_log_access {
	grant := data_adapter.trail_bucket_info.acl.Grants[_]
	not grant.Grantee.URI == "https://acs.amazonaws.com/groups/global/AllUsers"
	not grant.Grantee.URI == "https://acs.amazonaws.com/groups/global/AuthenticatedUsers"

	every statement in data_adapter.trail_bucket_info.policy.Statement {
		not statement.Effect == "Allow"
		not common.array_contains(["*", {"AWS": "*"}], statement.Principal)
	}
}
