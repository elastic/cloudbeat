package cis_eks.test_data

generate_eks_input(logging, encryption_config, endpoint_private_access, endpoint_public_access, public_access_cidrs) = {
	"type": "aws-eks",
	"resource": {"Cluster": {
		"Arn": "arn:aws:somearn1234:cluster/EKS-demo",
		"CertificateAuthority": {"Data": "some data"},
		"ClientRequestToken": null,
		"CreatedAt": "2021-10-27T11:08:51Z",
		"EncryptionConfig": encryption_config,
		"Endpoint": "https://C07EBEDB096B808626B023DDBF7520DC.gr7.us-east-2.eks.amazonaws.com",
		"Identity": {"Oidc": {"Issuer": "https://oidc.eks.us-east-2.amazonaws.com/id/C07EBdDB096B80AA626B023SS520SS"}},
		"Logging": logging,
		"ResourcesVpcConfig": {
			"ClusterSecurityGroupId": "sg-00abc463e0e831064",
			"EndpointPrivateAccess": endpoint_private_access,
			"EndpointPublicAccess": endpoint_public_access,
			"PublicAccessCidrs": public_access_cidrs,
			"SecurityGroupIds": ["sg-01f510f46974d3b5c"],
			"SubnetIds": [
				"subnet-03917f8779ce37c47",
				"subnet-09b8d7fdb5776ab47",
				"subnet-09021fed467f7ad25",
				"subnet-0885421a2d53b91c7",
			],
			"VpcId": "vpc-09b1bd8bbf4508a52",
		},
		"Name": "EKS-Elastic-agent-demo",
	}},
}

generate_eks_input_with_vpc_config(endpoint_private_access, endpoint_public_access, public_access_cidrs) = result {
	logging = {"ClusterLogging": [
		{
			"Enabled": false,
			"Types": [
				"authenticator",
				"controllerManager",
				"scheduler",
			],
		},
		{
			"Enabled": true,
			"Types": [
				"api",
				"audit",
			],
		},
	]}

	encryption_config = {"EncryptionConfig : null"}
	result = generate_eks_input(logging, encryption_config, endpoint_private_access, endpoint_public_access, public_access_cidrs)
}

generate_ecr_input_with_one_repo(image_scan_on_push) = {
	"type": "aws-ecr",
	"resource": {"EcrRepositories": [{
		"kind": "Pod",
		"CreatedAt": "2020-07-29T11:44:18Z",
		"ImageScanningConfiguration": {"ScanOnPush": image_scan_on_push},
		"ImageTagMutability": "MUTABLE",
		"RegistryId": "704479110758",
		"RepositoryArn": "arn:aws:ecr:us-east-2:704479110758:repository/build.security.management",
		"RepositoryName": "build.security.management",
		"RepositoryUri": "704479110758.dkr.ecr.us-east-2.amazonaws.com/build.security.management",
	}]},
}

generate_ecr_input_with_two_repo(first_image_scan_on_push, second_image_scan_on_push) = {
	"type": "aws-ecr",
	"resource": {"EcrRepositories": [
		{
			"kind": "Pod",
			"CreatedAt": "2020-07-29T11:44:18Z",
			"ImageScanningConfiguration": {"ScanOnPush": first_image_scan_on_push},
			"ImageTagMutability": "MUTABLE",
			"RegistryId": "704479110758",
			"RepositoryArn": "arn:aws:ecr:us-east-2:704479110758:repository/build.security.management",
			"RepositoryName": "build.security.management",
			"RepositoryUri": "704479110758.dkr.ecr.us-east-2.amazonaws.com/build.security.management",
		},
		{
			"CreatedAt": "2020-07-29T11:44:18Z",
			"ImageScanningConfiguration": {"ScanOnPush": second_image_scan_on_push},
			"ImageTagMutability": "MUTABLE",
			"RegistryId": "704479110758",
			"RepositoryArn": "arn:aws:ecr:us-east-2:704479110758:repository/build.security.management",
			"RepositoryName": "build.security.management2",
			"RepositoryUri": "704479110758.dkr.ecr.us-east-2.amazonaws.com/build.security.management",
		},
	]},
}

generate_elb_input_with_two_load_balancers(first_instance_protocol, first_instance_ssl_cert, sec_instance_protocol, sec_instance_ssl_cert) = {
	"type": "aws-elb",
	"resource": {"LoadBalancersDescription": [{
		"AvailabilityZones": [
			"us-east-2b",
			"us-east-2a",
		],
		"BackendServerDescriptions": null,
		"CanonicalHostedZoneName": "adda9cdc89b13452e92d48be46858d37-1423035038.us-east-2.elb.amazonaws.com",
		"CanonicalHostedZoneNameID": "Z3AADJGX6KTTL2",
		"CreatedTime": "2021-12-06T15:42:09.55Z",
		"DNSName": "adda9cdc89b13452e92d48be46858d37-1423035038.us-east-2.elb.amazonaws.com",
		"HealthCheck": {
			"HealthyThreshold": 2,
			"Interval": 10,
			"Target": "TCP:31829",
			"Timeout": 5,
			"UnhealthyThreshold": 6,
		},
		"Instances": [
			{"InstanceId": "i-0b81bd277b144e5e9"},
			{"InstanceId": "i-00e02dbffa8e2c54b"},
		],
		"ListenerDescriptions": [
			{
				"Listener": {
					"InstancePort": 32177,
					"InstanceProtocol": first_instance_protocol,
					"LoadBalancerPort": 443,
					"Protocol": "TCP",
					"SSLCertificateId": first_instance_ssl_cert,
				},
				"PolicyNames": null,
			},
			{
				"Listener": {
					"InstancePort": 31829,
					"InstanceProtocol": sec_instance_protocol,
					"LoadBalancerPort": 80,
					"Protocol": "TCP",
					"SSLCertificateId": sec_instance_ssl_cert,
				},
				"PolicyNames": null,
			},
		],
		"LoadBalancerName": "adda9cdc89b13452e92d48be46858d37",
		"Policies": {
			"AppCookieStickinessPolicies": null,
			"LBCookieStickinessPolicies": null,
			"OtherPolicies": null,
		},
		"Scheme": "internet-facing",
		"SecurityGroups": ["sg-08357d8bd7b80fc4c"],
		"SourceSecurityGroup": {
			"GroupName": "k8s-elb-adda9cdc89b13452e92d48be46858d37",
			"OwnerAlias": "704479110758",
		},
		"Subnets": [
			"subnet-09021fed467f7ad25",
			"subnet-09b8d7fdb5776ab47",
		],
		"VPCId": "vpc-09b1bd8bbf4508a52",
	}]},
}

not_evaluated_input = {
	"type": "some type",
	"resource": {"Cluster": {
		"Arn": "arn:aws:somearn1234:cluster/EKS-demo",
		"CertificateAuthority": {"Data": "some data"},
		"ClientRequestToken": null,
		"CreatedAt": "2021-10-27T11:08:51Z",
		"EncryptionConfig": null,
		"Endpoint": "https://C07EBEDB096B808626B023DDBF7520DC.gr7.us-east-2.eks.amazonaws.com",
		"Identity": {"Oidc": {"Issuer": "https://oidc.eks.us-east-2.amazonaws.com/id/C07EBdDB096B80AA626B023SS520SS"}},
		"Name": "EKS-Elastic-agent-demo",
	}},
}
