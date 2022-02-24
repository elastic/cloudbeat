package compliance.cis_eks.rules.cis_5_3_1

import data.compliance.cis_eks
import data.compliance.cis_eks.data_adatper
import data.compliance.lib.assert
import data.compliance.lib.common

default rule_evaluation = false

# Verify that there is a non empty encryption configuration
rule_evaluation {
	input.resource.Cluster.EncryptionConfig
	count(input.resource.Cluster.EncryptionConfig) > 0
}

# Ensure there Kuberenetes secrets are encrypted
finding = result {
	# filter
	data_adatper.is_aws_eks

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"encryption_config": input.resource.Cluster},
	}
}

metadata = {
	"name": "Ensure Kubernetes Secrets are encrypted using Customer Master Keys (CMKs) managed in AWS KMS",
	"description": "Encrypt Kubernetes secrets, stored in etcd, using secrets encryption feature during Amazon EKS cluster creation.",
	"rationale": `Kubernetes can store secrets that pods can access via a mounted volume.
Today, Kubernetes secrets are stored with Base64 encoding, but encrypting is the recommended approach.
Amazon EKS clusters version 1.13 and higher support the capability of encrypting your Kubernetes secrets using AWS Key Management Service (KMS) Customer Managed Keys (CMK).
The only requirement is to enable the encryption provider support during EKS cluster creation.
Use AWS Key Management Service (KMS) keys to provide envelope encryption of Kubernetes secrets stored in Amazon EKS. Implementing envelope encryption is considered a security best practice for applications that store sensitive data and is part of a defense in depth security strategy.
Application-layer Secrets Encryption provides an additional layer of security for sensitive data, such as user defined Secrets and Secrets required for the operation of the cluster, such as service account keys, which are all stored in etcd.
Using this functionality, you can use a key, that you manage in AWS KMS, to encrypt data at the application layer.
This protects against attackers in the event that they manage to gain access to etcd.`,
	"impact": `None`,
	"tags": array.concat(cis_eks.default_tags, ["CIS 5.3.1", "AWS Key Management Service (KMS)"]),
	"default_value": "By default, Application-layer Secrets Encryption is not enabled.",
	"benchmark": cis_eks.benchmark_metadata,
	"remediation": "Enable 'Secrets Encryption' during Amazon EKS cluster creation as described in the links within the 'References' section.",
}
