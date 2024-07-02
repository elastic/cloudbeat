package compliance.cis_aws.rules.cis_4_14

import data.compliance.lib.common
import data.compliance.policy.aws_cloudtrail.data_adapter
import data.compliance.policy.aws_cloudtrail.pattern
import data.compliance.policy.aws_cloudtrail.trail
import future.keywords.if

default rule_evaluation = false

finding = result if {
	# filter
	data_adapter.is_multi_trails_type

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(rule_evaluation),
		input.resource,
	)
}

# { ($.eventName = CreateVpc) || ($.eventName = DeleteVpc) || ($.eventName = ModifyVpcAttribute) || ($.eventName = AcceptVpcPeeringConnection) || ($.eventName = CreateVpcPeeringConnection) || ($.eventName = DeleteVpcPeeringConnection) || ($.eventName = RejectVpcPeeringConnection) || ($.eventName = AttachClassicLinkVpc) || ($.eventName = DetachClassicLinkVpc) || ($.eventName = DisableVpcClassicLink) || ($.eventName = EnableVpcClassicLink) }
required_patterns = [pattern.complex_expression("||", [
	pattern.simple_expression("$.eventName", "=", "CreateVpc"),
	pattern.simple_expression("$.eventName", "=", "DeleteVpc"),
	pattern.simple_expression("$.eventName", "=", "ModifyVpcAttribute"),
	pattern.simple_expression("$.eventName", "=", "AcceptVpcPeeringConnection"),
	pattern.simple_expression("$.eventName", "=", "CreateVpcPeeringConnection"),
	pattern.simple_expression("$.eventName", "=", "DeleteVpcPeeringConnection"),
	pattern.simple_expression("$.eventName", "=", "RejectVpcPeeringConnection"),
	pattern.simple_expression("$.eventName", "=", "AttachClassicLinkVpc"),
	pattern.simple_expression("$.eventName", "=", "DetachClassicLinkVpc"),
	pattern.simple_expression("$.eventName", "=", "DisableVpcClassicLink"),
	pattern.simple_expression("$.eventName", "=", "EnableVpcClassicLink"),
])]

rule_evaluation = trail.at_least_one_trail_satisfied(required_patterns)
