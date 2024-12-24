package compliance.policy.aws_cloudtrail.no_public_bucket_access

import data.compliance.policy.aws_cloudtrail.data_adapter
import future.keywords.if
import future.keywords.in

default bucket_is_public := false

# Bucket is public if any ACL grant grantee is `AllUsers`
bucket_is_public if {
	some grant in data_adapter.trail_bucket_info.acl.Grants
	grant.Grantee.URI == "http://acs.amazonaws.com/groups/global/AllUsers"
}

# Bucket is public if any ACL grant grantee is `AuthenticatedUsers`
bucket_is_public if {
	some grant in data_adapter.trail_bucket_info.acl.Grants
	grant.Grantee.URI == "http://acs.amazonaws.com/groups/global/AuthenticatedUsers"
}

# Bucket is public if any policy statement has effect "Allow" and principal "*"
bucket_is_public if {
	some statement in data_adapter.trail_bucket_info.policy.Statement
	statement.Effect == "Allow"
	statement.Principal == "*"
}

# Bucket is public if any policy statement has effect "Allow" and principal {"AWS": "*"}
bucket_is_public if {
	some statement in data_adapter.trail_bucket_info.policy.Statement
	statement.Effect == "Allow"
	statement.Principal == {"AWS": "*"}
}
