locals {
  windows_private_key_file = "${path.module}/cloudbeat-win-${random_id.id.hex}.pem"
  tags = merge({
    id          = random_id.id.hex
    provisioner = "terraform"
    Name        = var.deployment_name
  }, var.specific_tags)
}

resource "random_id" "id" {
  byte_length = 4
}

resource "tls_private_key" "cloudbeat_key" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "aws_key_pair" "generated_key" {
  provider   = aws
  key_name   = "cloudbeat-win-${random_id.id.hex}"
  public_key = tls_private_key.cloudbeat_key.public_key_openssh
  tags       = local.tags
}

data "aws_ami" "windows_2022" {
  count       = var.windows_ami_id == "" ? 1 : 0
  most_recent = true
  owners      = ["801119661308"]

  filter {
    name   = "name"
    values = ["Windows_Server-2022-English-Full-Base-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

resource "aws_security_group" "windows" {
  provider = aws

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "WinRM HTTP"
    from_port   = 5985
    to_port     = 5985
    protocol    = "tcp"
    cidr_blocks = [var.winrm_ingress_cidr]
  }

  tags = local.tags
}

resource "local_file" "cloud_pem" {
  filename        = local.windows_private_key_file
  content         = tls_private_key.cloudbeat_key.private_key_pem
  file_permission = "0400"
}

resource "aws_instance" "windows" {
  provider      = aws
  ami           = var.windows_ami_id != "" ? var.windows_ami_id : data.aws_ami.windows_2022[0].id
  instance_type = var.aws_ec2_instance_type
  key_name      = aws_key_pair.generated_key.key_name

  vpc_security_group_ids      = [aws_security_group.windows.id]
  associate_public_ip_address = true
  iam_instance_profile        = var.iam_instance_profile
  get_password_data           = true

  user_data = <<-EOT
    <powershell>
    $ErrorActionPreference = "Continue"
    try {
      Get-NetConnectionProfile | Where-Object { $_.IPv4Connectivity -eq "Internet" } | Set-NetConnectionProfile -NetworkCategory Private -ErrorAction SilentlyContinue
    } catch {}
    winrm quickconfig -q -force
    Enable-PSRemoting -Force -SkipNetworkProfileCheck
    Set-Item WSMan:\localhost\Service\AllowUnencrypted $true
    Set-Item WSMan:\localhost\Service\Auth\Basic $true
    Set-Item WSMan:\localhost\Client\AllowUnencrypted $true
    Set-Item WSMan:\localhost\Client\Auth\Basic $true
    netsh advfirewall firewall set rule group="Windows Remote Management" new enable=yes 2>$null
    Get-NetFirewallRule -DisplayGroup "Windows Remote Management" -ErrorAction SilentlyContinue | Where-Object { $_.Direction -eq "Inbound" } | Enable-NetFirewallRule -ErrorAction SilentlyContinue
    New-NetFirewallRule -DisplayName "WinRM-5985-CDR" -Direction Inbound -Protocol TCP -LocalPort 5985 -Action Allow -ErrorAction SilentlyContinue
    Set-Service WinRM -StartupType Automatic
    Restart-Service WinRM
    </powershell>
  EOT

  tags = local.tags

  depends_on = [local_file.cloud_pem]
}
