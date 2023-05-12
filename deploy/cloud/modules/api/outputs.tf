output "vanilla" {
  value = module.vanilla
}

output "eks" {
  value = module.eks
}

output "cspm_aws" {
  value = module.cspm_aws
}

output "installedCspm" {
  value = local.is8_7OrAbove
}

output "fleet_url" {
  value = local.fleet_url
}
