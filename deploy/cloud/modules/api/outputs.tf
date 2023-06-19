output "vanilla" {
  value = module.vanilla
}

output "eks" {
  value = module.eks
}

output "cspm_aws" {
  value = module.cspm_aws[0]
}

output "installedCspm" {
  value = local.is8_7OrAbove
}

output "fleet_url" {
  value = local.fleet_url
}
