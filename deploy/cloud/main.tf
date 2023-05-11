provider "ec" {
  apikey = var.ec_api_key
}

module "ec_deployment" {
  source = "github.com/elastic/apm-server/testing/infra/terraform/modules/ec_deployment"

  region        = var.ess_region
  stack_version = var.stack_version

  deployment_template    = var.deployment_template
  deployment_name_prefix = "${var.deployment_name}-${random_string.suffix.result}"

  integrations_server = true

  elasticsearch_size       = var.elasticsearch_size
  elasticsearch_zone_count = var.elasticsearch_zone_count

  docker_image = var.docker_image_override
  docker_image_tag_override = {
    "elasticsearch" : "",
    "kibana" : "",
    "apm" : ""
  }
}

module "eks" {
  source = "./modules/provision-eks-cluster"

  region       = var.eks_region
  cluster_name = var.deployment_name
}

# Retrieve EKS cluster information
provider "aws" {
  region = module.eks.region
}

data "aws_eks_cluster" "cluster" {
  name = module.eks.cluster_id
}

module "iam_eks_role" {
  source                 = "terraform-aws-modules/iam/aws//modules/iam-eks-role"
  depends_on             = [module.eks]
  allow_self_assume_role = true

  role_name = "Role-${var.deployment_name}-${random_string.suffix.result}"

  role_policy_arns = {
    Developers_eks = "arn:aws:iam::704479110758:policy/Developers_eks"
    EKS_ReadAccess = "arn:aws:iam::704479110758:policy/EKS_ReadAccess"
  }


  cluster_service_accounts = {
    (module.eks.cluster_name) = ["kube-system:elastic-agent"]
  }
}

data "local_file" "dashboard" {
  filename = "data/dashboard.ndjson"
}

resource "null_resource" "store_local_dashboard" {
  provisioner "local-exec" {
    command = "curl -X POST -u ${module.ec_deployment.elasticsearch_username}:${module.ec_deployment.elasticsearch_password} ${module.ec_deployment.kibana_url}/api/saved_objects/_import?overwrite=true -H \"kbn-xsrf: true\" --form file=@data/dashboard.ndjson"
  }
  depends_on = [module.ec_deployment]
  triggers = {
    dashboard_sha1 = sha1(file("data/dashboard.ndjson"))
  }
}


data "local_file" "rules" {
  filename = "data/rules.ndjson"
}

resource "null_resource" "rules" {
  provisioner "local-exec" {
    command = "curl -X POST -u ${module.ec_deployment.elasticsearch_username}:${module.ec_deployment.elasticsearch_password} ${module.ec_deployment.kibana_url}/api/saved_objects/_import?overwrite=true -H \"kbn-xsrf: true\" --form file=@data/rules.ndjson"
  }
  depends_on = [module.ec_deployment]
  triggers = {
    dashboard_sha1 = sha1(file("data/rules.ndjson"))
  }
}

provider "restapi" {
  username = module.ec_deployment.elasticsearch_username
  password = module.ec_deployment.elasticsearch_password
  uri      = module.ec_deployment.kibana_url

  debug                = true
  write_returns_object = true

  headers = {
    kbn-xsrf     = true
    content-type = "application/json"
  }

  # depends_on = [module.ec_deployment]
  # Currently this is not possible, this is why we need to apply multiple times with different targets.
  # See https://github.com/hashicorp/terraform/issues/2430 and https://github.com/Mastercard/terraform-provider-restapi/issues/20
}

module "api" {
  source = "./modules/api"

  providers  = { restapi : restapi }
  depends_on = [module.ec_deployment, module.iam_eks_role]

  username         = module.ec_deployment.elasticsearch_username
  password         = module.ec_deployment.elasticsearch_password
  uri              = module.ec_deployment.kibana_url
  role_arn         = module.iam_eks_role.iam_role_arn
  agent_docker_img = var.agent_docker_image_override
}

provider "kubernetes" {
  host                   = data.aws_eks_cluster.cluster.endpoint
  cluster_ca_certificate = base64decode(data.aws_eks_cluster.cluster.certificate_authority.0.data)
  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    args = [
      "eks",
      "get-token",
      "--cluster-name",
      data.aws_eks_cluster.cluster.name
    ]
  }
}

# In order for Elastic agent to successfully assume the role,
# it needs to be deployed (or restarted) after service account is created and annotated.
resource "kubernetes_manifest" "eks_agent_service_account" {
  depends_on = [module.eks, module.iam_eks_role, module.api.eks]
  for_each   = module.api.eks.service_account_manifests
  manifest   = each.value
}

resource "kubernetes_annotations" "eks_service_account" {
  api_version = "v1"
  kind        = "ServiceAccount"
  metadata {
    name      = "elastic-agent"
    namespace = "kube-system"
  }
  annotations = {
    "eks.amazonaws.com/role-arn" = module.iam_eks_role.iam_role_arn
  }
  depends_on = [kubernetes_manifest.eks_agent_service_account]
}

resource "kubernetes_manifest" "eks_agent_yaml" {
  depends_on = [kubernetes_annotations.eks_service_account]
  for_each   = module.api.eks.other_manifests
  manifest   = each.value
}


resource "random_string" "suffix" {
  length  = 3
  special = false
}

provider "helm" {
  kubernetes {
    host                   = data.aws_eks_cluster.cluster.endpoint
    cluster_ca_certificate = base64decode(data.aws_eks_cluster.cluster.certificate_authority.0.data)
    exec {
      api_version = "client.authentication.k8s.io/v1beta1"
      command     = "aws"
      args = [
        "eks",
        "get-token",
        "--cluster-name",
        data.aws_eks_cluster.cluster.name
      ]
    }
  }
}

module "apps" {
  source = "./modules/provision-apps"

  providers = {
    helm = helm
  }

  depends_on = [
    module.eks
  ]

  # nginx ingress replica count
  replica_count = "5"
}
module "aws_ec2_with_agent" {
  source    = "./modules/ec2"
  providers = { aws : aws }
  yml       = module.api.vanilla.yaml
  deployment_name = "${var.deployment_name}-${random_string.suffix.result}"
  depends_on = [
    module.ec_deployment,
    module.api,
  ]
}
