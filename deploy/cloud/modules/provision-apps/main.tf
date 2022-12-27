provider "helm" {
    kubernetes {
      config_path = pathexpand(var.kube_config)
    }
}
