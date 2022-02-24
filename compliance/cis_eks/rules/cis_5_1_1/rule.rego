package compliance.cis_eks.rules.cis_5_1_1

import data.compliance.cis_eks
import data.compliance.cis_eks.data_adatper
import data.compliance.lib.assert
import data.compliance.lib.common

default rule_evaluation = false

# Checks that every repository scanOnPush is enabled
rule_evaluation {
	input.resource.EcrRepositories

	# Verify there is no unsafe image
	misconfigured_repositories = [index | assert.is_false(input.resource.EcrRepositories[index].ImageScanningConfiguration.ScanOnPush)]
	count(misconfigured_repositories) == 0
}

evidence["misconfigured_repositories"] = misconfigured_repo {
	misconfigured_repo = [repo |
		repo := input.resource.EcrRepositories[index].RepositoryName
		assert.is_false(input.resource.EcrRepositories[index].ImageScanningConfiguration.ScanOnPush)
	]
}

# Check if image ScanOnPush is enabled
finding = result {
	# filter
	data_adatper.is_aws_ecr

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": evidence,
	}
}

metadata = {
	"name": "Ensure Image Vulnerability Scanning using Amazon ECR image scanning or a third party providers",
	"description": "Scan images being deployed to Amazon EKS for vulnerabilities.",
	"rationale": `Vulnerabilities in software packages can be exploited by hackers or malicious users to obtain unauthorized access to local cloud resources.
Amazon ECR and other third party products allow images to be scanned for known vulnerabilities.`,
	"remediation": `To utilize AWS ECR for Image scanning please follow the steps below:
To create a repository configured for scan on push (AWS CLI)

aws ecr create-repository --repository-name $REPO_NAME --image-scanning-configuration scanOnPush=true --region $REGION_CODE

To edit the settings of an existing repository (AWS CLI)

aws ecr put-image-scanning-configuration --repository-name $REPO_NAME --image-scanning-configuration scanOnPush=true --region $REGION_CODE
Use the following steps to start a manual image scan using the AWS Management Console.
1. Open the Amazon ECR console at https://console.aws.amazon.com/ecr/repositories.
2. From the navigation bar, choose the Region to create your repository in.
3. In the navigation pane, choose Repositories.
4. On the Repositories page, choose the repository that contains the image to scan.
5. On the Images page, select the image to scan and then choose Scan.`,
	"tags": array.concat(cis_eks.default_tags, ["CIS 5.1.1", "Image Registry and Image Scanning"]),
	"default_value": "Images are not scanned by Default.",
	"benchmark": cis_eks.benchmark_metadata,
	"impact": `If you are utilizing AWS ECR The following are common image scan failures.
You can view errors like this in the Amazon ECR console by displaying the image details or through the API or AWS CLI by using the DescribeImageScanFindings API.
UnsupportedImageError You may get an UnsupportedImageError error when attempting to scan an image that was built using an operating system that Amazon ECR doesn't support image scanning for.
Amazon ECR supports package vulnerability scanning for major versions of Amazon Linux, Amazon Linux 2, Debian, Ubuntu, CentOS, Oracle Linux, Alpine, and RHEL Linux distributions.
Amazon ECR does not support scanning images built from the Docker scratch image.
An UNDEFINED severity level is returned You may receive a scan finding that has a severity level of UNDEFINED.
The following are the common causes for this:
The vulnerability was not assigned a priority by the CVE source.
The vulnerability was assigned a priority that Amazon ECR did not recognize.
To determine the severity and description of a vulnerability, you can view the CVE directly from the source.`,
}
