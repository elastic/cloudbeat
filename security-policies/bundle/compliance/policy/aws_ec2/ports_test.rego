package compliance.policy.aws_ec2.ports

test_pass {
	in_range(80, 443, 443)
	in_range(80, 443, 80)
	in_range(80, 443, 100)
}

test_fail {
	not in_range(80, 443, 444)
	not in_range(80, 443, 30)
}
