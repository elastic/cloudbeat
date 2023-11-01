package compliance.cis_gcp.rules.cis_3_7

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test

type = "cloud-compute"

subtype = "gcp-compute-firewall"

test_violation {
	# specific port
	eval_fail with input as test_data.generate_gcp_asset(type, subtype, {"data": {"direction": "INGRESS", "sourceRanges": ["0.0.0.0/0"], "allowed": [{"IPProtocol": "tcp", "ports": ["3389"]}]}}, null)

	# port range
	eval_fail with input as test_data.generate_gcp_asset(type, subtype, {"data": {"direction": "INGRESS", "sourceRanges": ["0.0.0.0/0", "1.1.1.1/32"], "allowed": [{"IPProtocol": "tcp", "ports": ["3387-3400"]}, {"IPProtocol": "udp", "ports": ["40"]}]}}, null)

	# ALL protocols
	eval_fail with input as test_data.generate_gcp_asset(type, subtype, {"data": {"direction": "INGRESS", "sourceRanges": ["0.0.0.0/0"], "allowed": [{"IPProtocol": "all", "ports": ["3387-3400"]}]}}, null)

	# ALL protocols with no ports specified
	eval_fail with input as test_data.generate_gcp_asset(type, subtype, {"data": {"direction": "INGRESS", "sourceRanges": ["0.0.0.0/0"], "allowed": [{"IPProtocol": "all"}]}}, null)

	# TCP protocol with no ports specified, meaning ALL TCP ports are allowed
	eval_fail with input as test_data.generate_gcp_asset(type, subtype, {"data": {"direction": "INGRESS", "sourceRanges": ["0.0.0.0/0"], "allowed": [{"IPProtocol": "tcp"}]}}, null)
}

test_pass {
	eval_pass with input as test_data.generate_gcp_asset(type, subtype, {"data": {"direction": "INGRESS", "sourceRanges": ["0.0.0.0/0"], "allowed": [{"IPProtocol": "tcp", "ports": ["50"]}]}}, null)

	# source range is not open to the world
	eval_pass with input as test_data.generate_gcp_asset(type, subtype, {"data": {"direction": "INGRESS", "sourceRanges": ["1.1.1.1/32"], "allowed": [{"IPProtocol": "tcp", "ports": ["3389"]}]}}, null)

	# outbound traffic rule
	eval_pass with input as test_data.generate_gcp_asset(type, subtype, {"data": {"direction": "EGRESS", "sourceRanges": ["0.0.0.0/0"], "allowed": [{"IPProtocol": "tcp", "ports": ["3389"]}]}}, null)

	# protocol is not TCP or ALL
	eval_pass with input as test_data.generate_gcp_asset(type, subtype, {"data": {"direction": "INGRESS", "sourceRanges": ["0.0.0.0/0"], "allowed": [{"IPProtocol": "udp", "ports": ["3389"]}]}}, null)
}

test_not_evaluated {
	not_eval with input as test_data.not_eval_resource
}

eval_fail {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval {
	not finding with data.benchmark_data_adapter as data_adapter
}
