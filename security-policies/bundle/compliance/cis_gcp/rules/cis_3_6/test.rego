package compliance.cis_gcp.rules.cis_3_6

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test
import future.keywords.if

type := "cloud-compute"

subtype := "gcp-compute-firewall"

test_violation if {
	# specific port
	eval_fail with input as test_data.generate_gcp_asset(type, subtype, {"data": {"direction": "INGRESS", "sourceRanges": ["0.0.0.0/0"], "allowed": [{"IPProtocol": "tcp", "ports": ["22"]}]}}, null)

	# port range
	eval_fail with input as test_data.generate_gcp_asset(type, subtype, {"data": {"direction": "INGRESS", "sourceRanges": ["0.0.0.0/0", "1.1.1.1/32"], "allowed": [{"IPProtocol": "tcp", "ports": ["20-23"]}, {"IPProtocol": "udp", "ports": ["40"]}]}}, null)

	# ALL protocols
	eval_fail with input as test_data.generate_gcp_asset(type, subtype, {"data": {"direction": "INGRESS", "sourceRanges": ["0.0.0.0/0"], "allowed": [{"IPProtocol": "all", "ports": ["20-23"]}]}}, null)

	# ALL protocols with no ports specified
	eval_fail with input as test_data.generate_gcp_asset(type, subtype, {"data": {"direction": "INGRESS", "sourceRanges": ["0.0.0.0/0"], "allowed": [{"IPProtocol": "all"}]}}, null)

	# TCP protocol with no ports specified, meaning ALL TCP ports are allowed
	eval_fail with input as test_data.generate_gcp_asset(type, subtype, {"data": {"direction": "INGRESS", "sourceRanges": ["0.0.0.0/0"], "allowed": [{"IPProtocol": "tcp"}]}}, null)
}

test_pass if {
	eval_pass with input as test_data.generate_gcp_asset(type, subtype, {"data": {"direction": "INGRESS", "sourceRanges": ["0.0.0.0/0"], "allowed": [{"IPProtocol": "tcp", "ports": ["50"]}]}}, null)

	# source range is not open to the world
	eval_pass with input as test_data.generate_gcp_asset(type, subtype, {"data": {"direction": "INGRESS", "sourceRanges": ["1.1.1.1/32"], "allowed": [{"IPProtocol": "tcp", "ports": ["22"]}]}}, null)

	# outbound traffic rule
	eval_pass with input as test_data.generate_gcp_asset(type, subtype, {"data": {"direction": "EGRESS", "sourceRanges": ["0.0.0.0/0"], "allowed": [{"IPProtocol": "tcp", "ports": ["22"]}]}}, null)

	# protocol is not TCP or ALL
	eval_pass with input as test_data.generate_gcp_asset(type, subtype, {"data": {"direction": "INGRESS", "sourceRanges": ["0.0.0.0/0"], "allowed": [{"IPProtocol": "udp", "ports": ["22"]}]}}, null)
}

test_not_evaluated if {
	not_eval with input as test_data.not_eval_resource
}

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval if {
	not finding with data.benchmark_data_adapter as data_adapter
}
