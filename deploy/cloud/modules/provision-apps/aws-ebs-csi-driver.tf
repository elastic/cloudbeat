resource "helm_release" "aws_ebs_csi_driver" {
  chart = "aws-ebs-csi-driver"
  name = "aws-ebs-csi-driver"
  namespace = var.namespace
  repository = "https://kubernetes-sigs.github.io/aws-ebs-csi-driver"


  set {
    name = "controller.serviceAccount.name"
    value = "ebs-csi-controller-sa"
  }

}
