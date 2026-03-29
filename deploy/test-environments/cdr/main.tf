provider "aws" {
  region = var.region
}

provider "google" {
  project = var.gcp_project_id
}

provider "azurerm" {
  features {}
}

locals {
  common_tags = {
    division   = "${var.division}"
    org        = "${var.org}"
    team       = "${var.team}"
    project    = "${var.project}"
    owner      = "${var.owner}"
    deployment = "${var.deployment_name}"
  }

  deploy_gcp_vm              = var.cdr_elastic_defend_only ? false : var.deploy_gcp_vm
  deploy_az_vm               = var.cdr_elastic_defend_only ? false : var.deploy_az_vm
  deploy_aws_ec2             = var.cdr_elastic_defend_only ? false : var.deploy_aws_ec2
  deploy_aws_ec2_wiz         = var.cdr_elastic_defend_only ? false : var.deploy_aws_ec2_wiz
  deploy_aws_asset_inventory = var.cdr_elastic_defend_only ? false : var.deploy_aws_asset_inventory
}

resource "random_string" "suffix" {
  length  = 3
  special = false
}

# ===== CDR Infrastructure Resources =====

module "gcp_audit_logs" {
  count                    = local.deploy_gcp_vm ? 1 : 0
  providers                = { google : google }
  source                   = "../modules/gcp/vm"
  gcp_service_account_json = var.gcp_service_account_json
  deployment_name          = var.deployment_name
  network                  = "default"
  specific_tags            = merge(local.common_tags, { "vm_instance" : "audit-logs" })

}

module "aws_ec2_for_cloudtrail" {
  count           = local.deploy_aws_ec2 ? 1 : 0
  source          = "../modules/aws/ec2"
  providers       = { aws : aws }
  aws_ami         = var.ami_map[var.region]
  deploy_k8s      = false
  deploy_agent    = false
  deployment_name = "${var.deployment_name}-${random_string.suffix.result}"
  specific_tags   = merge(local.common_tags, { "ec2_type" : "cloudtrail" })
}

module "aws_ec2_for_wiz" {
  count           = local.deploy_aws_ec2_wiz ? 1 : 0
  source          = "../modules/aws/ec2"
  providers       = { aws : aws }
  aws_ami         = var.ami_map[var.region]
  deploy_k8s      = false
  deploy_agent    = false
  deployment_name = "${var.deployment_name}-${random_string.suffix.result}"
  specific_tags   = merge(local.common_tags, { "ec2_type" : "wiz" })
}

module "azure_vm_activity_logs" {
  count           = local.deploy_az_vm ? 1 : 0
  source          = "../modules/azure/vm"
  providers       = { azurerm : azurerm }
  location        = var.location
  deployment_name = var.deployment_name
  specific_tags   = merge(local.common_tags, { "vm_type" : "activity-logs" })
}

module "aws_ec2_for_asset_inventory" {
  count           = local.deploy_aws_asset_inventory ? 1 : 0
  source          = "../modules/aws/ec2"
  providers       = { aws : aws }
  aws_ami         = var.ami_map[var.region]
  deploy_k8s      = false
  deploy_agent    = false
  deployment_name = "${var.deployment_name}-${random_string.suffix.result}"
  specific_tags   = merge(local.common_tags, { "ec2_type" : "asset_inventory" })
}

module "aws_ec2_elastic_defend_linux" {
  count           = var.deploy_aws_elastic_defend_linux ? 1 : 0
  source          = "../modules/aws/ec2"
  providers       = { aws : aws }
  aws_ami         = var.ami_map[var.region]
  deploy_k8s      = false
  deploy_agent    = false
  deployment_name = "${var.deployment_name}-${random_string.suffix.result}"
  specific_tags   = merge(local.common_tags, { "ec2_type" : "elastic_defend_linux" })
}

module "aws_ec2_elastic_defend_windows" {
  count                 = var.deploy_aws_elastic_defend_windows ? 1 : 0
  source                = "../modules/aws/ec2-windows"
  providers             = { aws : aws }
  deployment_name       = "${var.deployment_name}-${random_string.suffix.result}"
  specific_tags         = merge(local.common_tags, { "ec2_type" : "elastic_defend_windows" })
  windows_ami_id        = var.windows_elastic_defend_ami_id
  aws_ec2_instance_type = var.windows_elastic_defend_instance_type
  winrm_ingress_cidr    = var.windows_elastic_defend_winrm_ingress_cidr
}
# ===== End Of CDR Infrastructure Resources =====
