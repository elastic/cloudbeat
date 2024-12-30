package compliance.policy.aws_elb.data_adapter

import future.keywords.if

is_aws_elb if {
	input.subType == "aws-elb"
}

cluster := input.resource.Cluster

listener_descriptions := input.resource.ListenerDescriptions

load_balancer_name := input.resource.LoadBalancerName
