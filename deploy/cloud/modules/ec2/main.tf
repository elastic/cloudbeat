locals {
  cloudbeat_private_key_file = "${path.module}/cloudbeat-${random_id.id.hex}.pem"
  ec2_username               = "ubuntu"
  tags = merge({
    id          = "${random_id.id.hex}"
    provisioner = "terraform"
    Name        = var.deployment_name
  }, var.specific_tags)
}

resource "tls_private_key" "cloudbeat_key" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "random_id" "id" {
  byte_length = 4
}

resource "aws_key_pair" "generated_key" {
  provider   = aws
  key_name   = "cloudbeat-generated-${random_id.id.hex}"
  public_key = tls_private_key.cloudbeat_key.public_key_openssh
  tags       = local.tags
}

resource "aws_security_group" "main" {
  provider = aws
  egress = [
    {
      cidr_blocks      = ["0.0.0.0/0", ]
      description      = ""
      from_port        = 0
      ipv6_cidr_blocks = []
      prefix_list_ids  = []
      protocol         = "-1"
      security_groups  = []
      self             = false
      to_port          = 0
    }
  ]
  ingress = [
    {
      cidr_blocks      = ["0.0.0.0/0", ]
      description      = ""
      from_port        = 22
      ipv6_cidr_blocks = []
      prefix_list_ids  = []
      protocol         = "tcp"
      security_groups  = []
      self             = false
      to_port          = 22
    }
  ]
  tags = local.tags

}


resource "local_file" "cloud_pem" {
  filename        = local.cloudbeat_private_key_file
  content         = tls_private_key.cloudbeat_key.private_key_pem
  file_permission = 0400
}


resource "aws_instance" "cloudbeat" {
  provider                    = aws
  ami                         = var.aws_ami
  instance_type               = var.aws_ec2_instance_type
  key_name                    = aws_key_pair.generated_key.key_name
  associate_public_ip_address = true
  vpc_security_group_ids      = [aws_security_group.main.id]
  iam_instance_profile        = "ec2-role-with-security-audit" # This is a prerequisite, role that contains the policy arn:aws:iam::aws:policy/SecurityAudit
  tags                        = local.tags
  connection {
    host        = self.public_ip
    user        = local.ec2_username
    private_key = tls_private_key.cloudbeat_key.private_key_pem
    timeout     = "2m"
  }

  provisioner "file" {
    content     = var.yml
    destination = "/tmp/manifests.yml"
  }
  provisioner "remote-exec" {
    inline = [
      "deploy_k8s=${var.deploy_k8s}",
      "if [ \"$deploy_k8s\" = true ]; then",
      "  echo 'Installing Kubernetes cluster using Kind tool'",
      "  cloud-init status --wait",
      "  git clone https://github.com/elastic/cloudbeat",
      "  cd cloudbeat",
      "  sudo kind create cluster --config deploy/k8s/kind/kind-multi.yml --wait 30s",
      "  sudo kind export kubeconfig --name kind-multi --kubeconfig /home/ubuntu/.kube/config",
      "  enable_agent=${var.deploy_agent}",
      "  if [ \"$enable_agent\" = true ]; then",
      "    echo 'Deploy KSPM agent'",
      "    kubectl apply -f /tmp/manifests.yml",
      "    ${var.cspm_aws_docker_cmd}",
      "  else",
      "    echo 'KSPM Agent will not be installed!'",
      "  fi",
      "else",
      "  echo 'No Kubernetes cluster will be installed'",
      "fi"
    ]
  }
}
