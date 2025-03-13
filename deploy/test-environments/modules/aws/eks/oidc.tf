resource "aws_eks_identity_provider_config" "eks_oidc" {
  cluster_name = local.cluster_name
  oidc {
    client_id                     = module.eks.oidc_provider
    identity_provider_config_name = "${local.cluster_name}-oidc"
    issuer_url                    = module.eks.cluster_oidc_issuer_url

  }

  depends_on = [module.eks, null_resource.wait_for_cluster]
}

resource "null_resource" "wait_for_cluster" {
  depends_on = [module.eks]

  provisioner "local-exec" {
    command = <<-EOT
      aws eks wait cluster-active --name ${local.cluster_name} --region ${var.region}
    EOT
  }
}
