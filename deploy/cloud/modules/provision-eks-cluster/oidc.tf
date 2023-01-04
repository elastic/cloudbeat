resource "aws_eks_identity_provider_config" "eks_oidc" {
  cluster_name = local.cluster_name
  oidc {
    client_id                     = module.eks.oidc_provider
    identity_provider_config_name = "${var.cluster_name_prefix}-oidc"
    issuer_url                    = module.eks.cluster_oidc_issuer_url

  }

  depends_on = [module.eks]
}
