package compliance.policy.aws_ec2.data_adapter

is_nacl_policy {
	input.subType == "aws-nacl"
}

nacl_entries = entries {
	entries := input.resource.Entries
}

# Filter all the entries that 
# 1. have ingres (egress == false)
# 2. allow any source ip of 0.0.0.0/0
nacl_ingresses = res {
	res = [entry | entry := nacl_entries[_]; entry.Egress == false; entry.CidrBlock == "0.0.0.0/0"; entry.RuleAction == "allow"]
}

# If the PortRange field is not specified for a network ACL rule, 
# it means that the rule applies to all ports for the specified protocol. 
# For example, if you create a rule that allows inbound traffic on TCP protocol and do not specify a PortRange, 
# the rule will allow inbound traffic on all TCP ports.
ingresses_with_all_ports_open = res {
	res = [entry | entry := nacl_ingresses[_]; not entry.PortRange]
}
