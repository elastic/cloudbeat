module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "18.26.6"

  cluster_name    = local.cluster_name
  cluster_version = "1.24"

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  enable_irsa                 = true
  create_cloudwatch_log_group = false

  tags = var.tags

  eks_managed_node_group_defaults = {
    ami_type = "AL2_x86_64"

    attach_cluster_primary_security_group = true

    # Disabling and using externally provided security groups
    create_security_group = false
  }

  eks_managed_node_groups = var.enable_node_group_two ? {
    one = {
      name = "${local.cluster_name}-1"

      instance_types = ["t3.small"]

      min_size     = 1
      max_size     = 4
      desired_size = var.node_group_one_desired_size

      vpc_security_group_ids = [
        aws_security_group.node_group_one.id
      ]
    },
    two = {
      name = "${local.cluster_name}-2"

      instance_types = ["t3.medium"]

      min_size     = 1
      max_size     = 4
      desired_size = var.node_group_two_desired_size

      vpc_security_group_ids = [
        aws_security_group.node_group_two.id
      ]
    }
    } : {
    one = {
      name = "${var.cluster_name}-1"

      instance_types = ["t3.small"]

      min_size     = 1
      max_size     = 4
      desired_size = var.node_group_one_desired_size

      vpc_security_group_ids = [
        aws_security_group.node_group_one.id
      ]
    }
  }

  manage_aws_auth_configmap = true

  aws_auth_roles = [
    {
      groups = [
        "system:masters",
      ]
      rolearn = "arn:aws:iam::704479110758:role/Developer_eks"
    }
  ]
}
