package compliance.cis_aws.rules.cis_4_14

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.compliance.policy.aws_cloudtrail.pattern
import data.lib.test
import future.keywords.if

test_pass if {
	eval_pass with input as rule_input([{
		"TrailInfo": {
			"Trail": {"IsMultiRegionTrail": true},
			"Status": {"IsLogging": true},
			"EventSelectors": [{"IncludeManagementEvents": true, "ReadWriteType": "All"}],
		},
		"MetricFilters": [{
			"FilterName": "filter_1",
			"FilterPattern": "{ ($.eventName = CreateVpc) || ($.eventName = DeleteVpc) || ($.eventName = ModifyVpcAttribute) || ($.eventName = AcceptVpcPeeringConnection) || ($.eventName = CreateVpcPeeringConnection) || ($.eventName = DeleteVpcPeeringConnection) || ($.eventName = RejectVpcPeeringConnection) || ($.eventName = AttachClassicLinkVpc) || ($.eventName = DetachClassicLinkVpc) || ($.eventName = DisableVpcClassicLink) || ($.eventName = EnableVpcClassicLink) }",
			"ParsedFilterPattern": pattern.complex_expression("||", [
				pattern.simple_expression("$.eventName", "=", "RejectVpcPeeringConnection"),
				pattern.simple_expression("$.eventName", "=", "DeleteVpc"),
				pattern.simple_expression("$.eventName", "=", "ModifyVpcAttribute"),
				pattern.simple_expression("$.eventName", "=", "CreateVpcPeeringConnection"),
				pattern.simple_expression("$.eventName", "=", "DisableVpcClassicLink"),
				pattern.simple_expression("$.eventName", "=", "CreateVpc"),
				pattern.simple_expression("$.eventName", "=", "AttachClassicLinkVpc"),
				pattern.simple_expression("$.eventName", "=", "DetachClassicLinkVpc"),
				pattern.simple_expression("$.eventName", "=", "DeleteVpcPeeringConnection"),
				pattern.simple_expression("$.eventName", "=", "AcceptVpcPeeringConnection"),
				pattern.simple_expression("$.eventName", "=", "EnableVpcClassicLink"),
			]),
		}],
		"MetricTopicBinding": {"filter_1": ["arn:aws:...sns"]},
	}])
}

rule_input(entry) := test_data.generate_monitoring_resources(entry)

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}
