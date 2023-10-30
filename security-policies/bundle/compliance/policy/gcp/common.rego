package compliance.policy.gcp.common

# parse the machine's family type from a machine type URL (e.g. https://www.googleapis.com/compute/v1/projects/<PROJECT_ID>/zones/<ZONE>/machineTypes/<FAMILY_TYPE>)
get_machine_type_family(type_url) = family {
	parts := split(type_url, "/")
	family := parts[count(parts) - 1]
}
