package compliance.policy.aws_elb.data_adapter

is_aws_elb {
	input.subType == "aws-elb"
}

cluster = input.resource.Cluster

listener_descriptions = input.resource.ListenerDescriptions

load_balancer_name = input.resource.LoadBalancerName
