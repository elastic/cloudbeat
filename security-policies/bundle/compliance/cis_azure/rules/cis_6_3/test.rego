package compliance.cis_azure.rules.cis_6_3

import data.cis_azure.test_data
import data.compliance.policy.azure.data_adapter
import data.lib.test
import future.keywords.if

test_pass if {
	# undefined extension
	eval_pass with input as test_data.generate_vm({})

	# empty rules
	eval_pass with input as test_data.generate_vm_with_extension({"network": {"securityRules": []}})

	assert_udp_pass("53")
	assert_udp_pass("123")
	assert_udp_pass("161")
	assert_udp_pass("389")
	assert_udp_pass("1900")
}

test_fail if {
	assert_udp_fail("53")
	assert_udp_fail("123")
	assert_udp_fail("161")
	assert_udp_fail("389")
	assert_udp_fail("1900")
}

test_not_evaluated if {
	not_eval with input as test_data.not_eval_non_exist_type
	not_eval with input as test_data.not_eval_non_exist_type
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

# regal ignore:rule-length
assert_udp_pass(port) if {
	# 1 valid rule
	eval_pass with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": "8080",
		"destinationPortRanges": [],
		"direction": "Inbound",
		"protocol": "UDP",
		"sourceAddressPrefix": "internet",
		"sourceAddressPrefixes": [],
	}]}})

	# Valid block
	eval_pass with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"destinationPortRange": port,
		"protocol": "UDP",
		"access": "Block",
		"destinationPortRanges": [],
		"direction": "Inbound",
		"sourceAddressPrefix": "internet",
		"sourceAddressPrefixes": [],
	}]}})

	# Valid block TCP
	eval_pass with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": port,
		"destinationPortRanges": [],
		"direction": "Inbound",
		"protocol": "TCP",
		"sourceAddressPrefix": "internet",
		"sourceAddressPrefixes": [],
	}]}})

	# Valid allow for specific ip
	eval_pass with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"destinationPortRange": port,
		"protocol": "UDP",
		"access": "Allow",
		"destinationPortRanges": [],
		"direction": "Inbound",
		"sourceAddressPrefix": "156.178.100.87",
		"sourceAddressPrefixes": [],
	}]}})

	# Valid range
	eval_pass with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"destinationPortRange": sprintf("%s-65365", [to_number(port) + 1]),
		"protocol": "UDP",
		"access": "Allow",
		"destinationPortRanges": [],
		"direction": "Inbound",
		"sourceAddressPrefix": "internet",
		"sourceAddressPrefixes": [],
	}]}})

	# Valid allow Outbound
	eval_pass with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"destinationPortRange": port,
		"protocol": "UDP",
		"access": "Allow",
		"destinationPortRanges": [],
		"direction": "Outbound",
		"sourceAddressPrefix": "internet",
		"sourceAddressPrefixes": [],
	}]}})

	eval_pass with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"destinationPortRanges": ["20", "30,12", "11,17,20-22"],
		"protocol": "UDP",
		"access": "Allow",
		"destinationPortRange": "",
		"direction": "Inbound",
		"sourceAddressPrefix": "internet",
		"sourceAddressPrefixes": [],
	}]}})

	# Valid allow without bad source
	eval_pass with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": port,
		"destinationPortRanges": [],
		"direction": "Inbound",
		"protocol": "UDP",
		"sourceAddressPrefix": "",
		"sourceAddressPrefixes": ["156.198.196.12", "100.0.0.0"],
	}]}})

	# 2 valid rule
	eval_pass with input as test_data.generate_vm_with_extension({"network": {"securityRules": [
		{
			"access": "Allow",
			"destinationPortRange": "8080",
			"destinationPortRanges": [],
			"direction": "Inbound",
			"protocol": "UDP",
			"sourceAddressPrefix": "internet",
			"sourceAddressPrefixes": [],
		},
		{
			"access": "Block",
			"destinationPortRange": "3389",
			"destinationPortRanges": [],
			"direction": "Outbound",
			"protocol": "UDP",
			"sourceAddressPrefix": "internet",
			"sourceAddressPrefixes": [],
		},
	]}})
}

# regal ignore:rule-length
assert_udp_fail(port) if {
	# Fail with port
	eval_fail with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": port,
		"destinationPortRanges": [],
		"direction": "Inbound",
		"protocol": "UDP",
		"sourceAddressPrefix": "internet",
		"sourceAddressPrefixes": [],
	}]}})

	# Fail with port and range
	eval_fail with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": sprintf("%s,76-80", [port]),
		"destinationPortRanges": [],
		"direction": "Inbound",
		"protocol": "UDP",
		"sourceAddressPrefix": "internet",
		"sourceAddressPrefixes": [],
	}]}})

	# Fail with port as lower range boundary
	eval_fail with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": sprintf("%s-60000", [port]),
		"destinationPortRanges": [],
		"direction": "Inbound",
		"protocol": "UDP",
		"sourceAddressPrefix": "internet",
		"sourceAddressPrefixes": [],
	}]}})

	# Fail with port as upper range boundary
	eval_fail with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": sprintf("10-%s", [port]),
		"destinationPortRanges": [],
		"direction": "Inbound",
		"protocol": "UDP",
		"sourceAddressPrefix": "internet",
		"sourceAddressPrefixes": [],
	}]}})

	# Fail with port is in range
	eval_fail with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": "10-60000",
		"destinationPortRanges": [],
		"direction": "Inbound",
		"protocol": "UDP",
		"sourceAddressPrefix": "internet",
		"sourceAddressPrefixes": [],
	}]}})

	# Fail with port is in ranges
	eval_fail with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": "",
		"destinationPortRanges": ["80", sprintf("10-%s", [port]), "60000-60200"],
		"direction": "Inbound",
		"protocol": "UDP",
		"sourceAddressPrefix": "internet",
		"sourceAddressPrefixes": [],
	}]}})

	# Fail with port and source address any in prefixes
	eval_fail with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": port,
		"destinationPortRanges": [],
		"direction": "Inbound",
		"protocol": "UDP",
		"sourceAddressPrefix": "",
		"sourceAddressPrefixes": ["197.198.158.2", "any"],
	}]}})

	# Fail with port and source address any
	eval_fail with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": port,
		"destinationPortRanges": [],
		"direction": "Inbound",
		"protocol": "UDP",
		"sourceAddressPrefix": "any",
		"sourceAddressPrefixes": [],
	}]}})

	# Fail with port and source address <nw>/0
	eval_fail with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": port,
		"destinationPortRanges": [],
		"direction": "Inbound",
		"protocol": "UDP",
		"sourceAddressPrefix": "<nw>/0",
		"sourceAddressPrefixes": [],
	}]}})

	# Fail with port and source address 0.0.0.0
	eval_fail with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": port,
		"destinationPortRanges": [],
		"direction": "Inbound",
		"protocol": "UDP",
		"sourceAddressPrefix": "0.0.0.0",
		"sourceAddressPrefixes": [],
	}]}})

	# Fail with port and source address *
	eval_fail with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": port,
		"destinationPortRanges": [],
		"direction": "Inbound",
		"protocol": "UDP",
		"sourceAddressPrefix": "*",
		"sourceAddressPrefixes": [],
	}]}})
}
