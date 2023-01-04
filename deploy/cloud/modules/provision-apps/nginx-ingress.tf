resource "helm_release" "nginx_ingress" {
  chart      = "nginx-ingress-controller"
  name       = "nginx-ingress-controller"

  repository = "https://charts.bitnami.com/bitnami"
  timeout = 600
  namespace = var.namespace

  set {
    name  = "service.type"
    value = "ClusterIP"
  }

  set {
    name = "replicaCount"
    value = var.replica_count
  }
}
