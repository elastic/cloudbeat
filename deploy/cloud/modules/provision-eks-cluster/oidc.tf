resource "aws_eks_identity_provider_config" "demo" {
  cluster_name = local.cluster_name
  oidc {
    client_id                     = module.eks.oidc_provider
    identity_provider_config_name = "cloudbeat-tf-oidc"
    issuer_url                    = module.eks.cluster_oidc_issuer_url

  }

  depends_on = [module.eks]
}
