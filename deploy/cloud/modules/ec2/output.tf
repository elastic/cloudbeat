output "aws_instance_cloudbeat_public_ip" {
  description = "AWS EC2 instance public IP"
  value       = aws_instance.cloudbeat.public_ip
}

output "cloudbeat_ssh_cmd" {
  description = "Use this command SSH into the ec2 instance"
  value       = "ssh -i ${local.cloudbeat_private_key_file} ${local.ec2_username}@${aws_instance.cloudbeat.public_ip}"
}

output "ec2_private_key" {
  description = "Use this private key to SSH into the ec2 instance"
  value = tls_private_key.cloudbeat_key.private_key_pem
  sensitive = true
}
