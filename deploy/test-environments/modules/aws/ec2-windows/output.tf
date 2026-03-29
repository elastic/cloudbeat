output "aws_instance_public_ip" {
  description = "Windows EC2 public IP"
  value       = aws_instance.windows.public_ip
}

output "ec2_ssh_key" {
  description = "Path to PEM for the generated instance key pair"
  value       = local.windows_private_key_file
}

output "instance_id" {
  description = "EC2 instance id (for get-password-data)"
  value       = aws_instance.windows.id
}
