package compliance.cis_azure.rules.cis_6_2

import data.cis_azure.test_data
import data.compliance.policy.azure.data_adapter
import data.lib.test
import future.keywords.if

# regal ignore:rule-length
test_pass if {
	# undefined extension
	eval_pass with input as test_data.generate_vm({})

	# empty rules
	eval_pass with input as test_data.generate_vm_with_extension({"network": {"securityRules": []}})

	# 1 valid rule
	eval_pass with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": "8080",
		"destinationPortRanges": [],
		"direction": "Inbound",
		"protocol": "TCP",
		"sourceAddressPrefix": "internet",
		"sourceAddressPrefixes": [],
	}]}})

	# Valid block 22
	eval_pass with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Block",
		"destinationPortRange": "22",
		"destinationPortRanges": [],
		"direction": "Inbound",
		"protocol": "TCP",
		"sourceAddressPrefix": "internet",
		"sourceAddressPrefixes": [],
	}]}})

	# Valid block 22 UDP
	eval_pass with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": "22",
		"destinationPortRanges": [],
		"direction": "Inbound",
		"protocol": "UDP",
		"sourceAddressPrefix": "internet",
		"sourceAddressPrefixes": [],
	}]}})

	# Valid allow 22 for specific ip
	eval_pass with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": "22",
		"destinationPortRanges": [],
		"direction": "Inbound",
		"protocol": "TCP",
		"sourceAddressPrefix": "156.178.100.87",
		"sourceAddressPrefixes": [],
	}]}})

	# Valid range
	eval_pass with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": "23-80",
		"destinationPortRanges": [],
		"direction": "Inbound",
		"protocol": "TCP",
		"sourceAddressPrefix": "internet",
		"sourceAddressPrefixes": [],
	}]}})

	# Valid allow 22 Outbound
	eval_pass with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": "22",
		"destinationPortRanges": [],
		"direction": "Outbound",
		"protocol": "TCP",
		"sourceAddressPrefix": "internet",
		"sourceAddressPrefixes": [],
	}]}})

	# Valid allow 22 not in ranges
	eval_pass with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": "",
		"destinationPortRanges": ["20", "30,12", "11,17,49-60"],
		"direction": "Inbound",
		"protocol": "TCP",
		"sourceAddressPrefix": "internet",
		"sourceAddressPrefixes": [],
	}]}})

	# Valid allow 22 without bad source
	eval_pass with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": "22",
		"destinationPortRanges": [],
		"direction": "Inbound",
		"protocol": "TCP",
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
			"protocol": "TCP",
			"sourceAddressPrefix": "internet",
			"sourceAddressPrefixes": [],
		},
		{
			"access": "Block",
			"destinationPortRange": "22",
			"destinationPortRanges": [],
			"direction": "Outbound",
			"protocol": "TCP",
			"sourceAddressPrefix": "internet",
			"sourceAddressPrefixes": [],
		},
	]}})
}

# regal ignore:rule-length
test_fail if {
	# Fail with port 22
	eval_fail with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": "22",
		"destinationPortRanges": [],
		"direction": "Inbound",
		"protocol": "TCP",
		"sourceAddressPrefix": "internet",
		"sourceAddressPrefixes": [],
	}]}})

	# Fail with port 22 and range
	eval_fail with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": "22,76-80",
		"destinationPortRanges": [],
		"direction": "Inbound",
		"protocol": "TCP",
		"sourceAddressPrefix": "internet",
		"sourceAddressPrefixes": [],
	}]}})

	# Fail with port 22 as lower range boundary
	eval_fail with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": "22-30",
		"destinationPortRanges": [],
		"direction": "Inbound",
		"protocol": "TCP",
		"sourceAddressPrefix": "internet",
		"sourceAddressPrefixes": [],
	}]}})

	# Fail with port 22 as upper range boundary
	eval_fail with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": "10-22",
		"destinationPortRanges": [],
		"direction": "Inbound",
		"protocol": "TCP",
		"sourceAddressPrefix": "internet",
		"sourceAddressPrefixes": [],
	}]}})

	# Fail with port 22 is in range
	eval_fail with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": "10-30",
		"destinationPortRanges": [],
		"direction": "Inbound",
		"protocol": "TCP",
		"sourceAddressPrefix": "internet",
		"sourceAddressPrefixes": [],
	}]}})

	# Fail with port 22 is in ranges
	eval_fail with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": "",
		"destinationPortRanges": ["80", "22-37", "3300-3400"],
		"direction": "Inbound",
		"protocol": "TCP",
		"sourceAddressPrefix": "internet",
		"sourceAddressPrefixes": [],
	}]}})

	# Fail with port 22 and source address any in prefixes
	eval_fail with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": "22",
		"destinationPortRanges": [],
		"direction": "Inbound",
		"protocol": "TCP",
		"sourceAddressPrefix": "",
		"sourceAddressPrefixes": ["197.198.158.2", "any"],
	}]}})

	# Fail with port 22 and source address any
	eval_fail with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": "22",
		"destinationPortRanges": [],
		"direction": "Inbound",
		"protocol": "TCP",
		"sourceAddressPrefix": "any",
		"sourceAddressPrefixes": [],
	}]}})

	# Fail with port 22 and source address <nw>/0
	eval_fail with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": "22",
		"destinationPortRanges": [],
		"direction": "Inbound",
		"protocol": "TCP",
		"sourceAddressPrefix": "<nw>/0",
		"sourceAddressPrefixes": [],
	}]}})

	# Fail with port 22 and source address 0.0.0.0
	eval_fail with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": "22",
		"destinationPortRanges": [],
		"direction": "Inbound",
		"protocol": "TCP",
		"sourceAddressPrefix": "0.0.0.0",
		"sourceAddressPrefixes": [],
	}]}})

	# Fail with port 22 and source address *
	eval_fail with input as test_data.generate_vm_with_extension({"network": {"securityRules": [{
		"access": "Allow",
		"destinationPortRange": "22",
		"destinationPortRanges": [],
		"direction": "Inbound",
		"protocol": "TCP",
		"sourceAddressPrefix": "*",
		"sourceAddressPrefixes": [],
	}]}})
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
