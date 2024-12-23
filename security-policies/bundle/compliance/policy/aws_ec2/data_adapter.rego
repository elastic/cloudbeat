package compliance.policy.aws_ec2.data_adapter

import future.keywords.if

is_nacl_policy if {
	input.subType == "aws-nacl"
}

is_security_group_policy if {
	input.subType == "aws-security-group"
}

is_vpc_policy if {
	input.subType == "aws-vpc"
}

is_ebs_policy if {
	input.subType == "aws-ebs"
}

nacl_entries := input.resource.Entries

security_groups_ip_permissions := input.resource.IpPermissions

is_default_security_group if {
	input.resource.GroupName == "default"
}

# Filter all the entries that
# 1. have ingres (egress == false)
# 2. allow any source ip of 0.0.0.0/0
nacl_ingresses := [entry | entry := nacl_entries[_]; entry.Egress == false; entry.CidrBlock == "0.0.0.0/0"; entry.RuleAction == "allow"]

# If the PortRange field is not specified for a network ACL rule,
# it means that the rule applies to all ports for the specified protocol.
# For example, if you create a rule that allows inbound traffic on TCP protocol and do not specify a PortRange,
# the rule will allow inbound traffic on all TCP ports.
ingresses_with_all_ports_open := [entry | entry := nacl_ingresses[_]; not entry.PortRange]

# all the IpRanges from security groups that has an open inbound for all ipv4 cidr notions
public_ipv4 := [entry | entry := security_groups_ip_permissions[_]; entry.IpRanges[_].CidrIp == "0.0.0.0/0"]

# all the IpRangesv6 from security groups that has an open inbound for all ipv6 cidr notions
public_ipv6 := [entry | entry := security_groups_ip_permissions[_]; entry.Ipv6Ranges[_].CidrIpv6 == "::/0"]

security_group_inbound_rules := input.resource.IpPermissions

security_group_outbound_rules := input.resource.IpPermissionsEgress
