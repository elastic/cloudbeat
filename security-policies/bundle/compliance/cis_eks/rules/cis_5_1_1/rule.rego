package compliance.cis_eks.rules.cis_5_1_1

import data.compliance.policy.aws_ecr.ensure_image_scan as audit

# Check if image ScanOnPush is enabled
finding := audit.finding
