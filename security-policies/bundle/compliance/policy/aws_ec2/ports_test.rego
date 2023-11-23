package compliance.policy.aws_ec2.ports

import future.keywords.if

test_pass if {
	in_range(80, 443, 443)
	in_range(80, 443, 80)
	in_range(80, 443, 100)
}

test_fail if {
	not in_range(80, 443, 444)
	not in_range(80, 443, 30)
}
