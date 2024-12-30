package compliance.cis_eks.rules.cis_5_4_3

import data.cis_eks.test_data as eks_test_data
import data.kubernetes_common.test_data
import data.lib.test
import future.keywords.if

test_violation if {
	test.assert_fail(finding) with input as rule_input(violating_input_public_ip_and_public_address)
}

test_pass if {
	test.assert_pass(finding) with input as rule_input(valid_input_no_external_IP)
	test.assert_pass(finding) with input as rule_input(valid_input_external_IP_set_to_local_host)
}

test_not_evaluated if {
	not finding with input as eks_test_data.not_evaluated_input
}

rule_input(resource) := test_data.kube_api_input(resource)

violating_input_public_ip_and_public_address := {
	"kind": "Node",
	"status": {"addresses": [
		{
			"type": "InternalIP",
			"address": "192.168.54.45",
		},
		{
			"type": "ExternalIP",
			"address": "18.119.116.97",
		},
		{
			"type": "Hostname",
			"address": "ip-192-168-54-45.us-east-2.compute.internal",
		},
		{
			"type": "InternalDNS",
			"address": "ip-192-168-54-45.us-east-2.compute.internal",
		},
		{
			"type": "ExternalDNS",
			"address": "ec2-18-119-116-97.us-east-2.compute.amazonaws.com",
		},
	]},
}

valid_input_no_external_IP := {
	"kind": "Node",
	"status": {"addresses": [
		{
			"type": "InternalIP",
			"address": "192.168.54.45",
		},
		{
			"type": "Hostname",
			"address": "ip-192-168-54-45.us-east-2.compute.internal",
		},
		{
			"type": "InternalDNS",
			"address": "ip-192-168-54-45.us-east-2.compute.internal",
		},
		{
			"type": "ExternalDNS",
			"address": "ec2-18-119-116-97.us-east-2.compute.amazonaws.com",
		},
	]},
}

valid_input_external_IP_set_to_local_host := {
	"kind": "Node",
	"status": {"addresses": [
		{
			"type": "InternalIP",
			"address": "192.168.54.45",
		},
		{
			"type": "ExternalIP",
			"address": "0.0.0.0",
		},
		{
			"type": "Hostname",
			"address": "ip-192-168-54-45.us-east-2.compute.internal",
		},
		{
			"type": "InternalDNS",
			"address": "ip-192-168-54-45.us-east-2.compute.internal",
		},
		{
			"type": "ExternalDNS",
			"address": "ec2-18-119-116-97.us-east-2.compute.amazonaws.com",
		},
	]},
}
