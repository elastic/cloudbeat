package compliance.policy.aws_ec2.ports

import future.keywords.if

# Admin ports are network ports that are reserved for use by system administrators to manage servers and other network devices.
# These ports are typically used for remote management, monitoring, and control of devices over a network
admin_ports := {22, 23, 25, 53, 80, 110, 143, 389, 443, 465, 587, 636, 993, 995, 3389}

# check whether a given value (candidate) is within a range of values specified by from and to
in_range(from, to, candidate) if {
	candidate >= from
	candidate <= to
} else := false
