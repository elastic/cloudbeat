output "elasticsearch_url" {
  value       = module.ec_deployment.elasticsearch_url
  description = "The secure Elasticsearch URL"
}

output "elasticsearch_username" {
  value       = module.ec_deployment.elasticsearch_username
  description = "The Elasticsearch username"
  sensitive   = true
}

output "elasticsearch_password" {
  value       = module.ec_deployment.elasticsearch_password
  description = "The Elasticsearch password"
  sensitive   = true
}

output "kibana_url" {
  value       = module.ec_deployment.kibana_url
  description = "The secure Kibana URL"
}

output "admin_console_url" {
  value       = module.ec_deployment.admin_console_url
  description = "The admin console URL"
}

output "eks_cluster_id" {
  description = "EKS cluster ID"
  value       = module.eks.cluster_id
}

output "eks_cluster_endpoint" {
  description = "Endpoint for EKS control plane"
  value       = module.eks.cluster_endpoint
}

output "eks_cluster_security_group_id" {
  description = "Security group ids attached to the cluster control plane"
  value       = module.eks.cluster_security_group_id
}

output "eks_region" {
  description = "AWS region"
  value       = module.eks.region
}

output "eks_cluster_name" {
  description = "Kubernetes Cluster Name"
  value       = module.eks.cluster_name
}

output "agent_policy_id" {
  description = "Agent policy ID"
  value       = module.api.agent_policy_id
}

output "enrollment_token" {
  description = "Agent enrollment token"
  value       = module.api.enrollment_token
}

output "fleet_url" {
  description = "Fleet Server URL"
  value       = module.api.fleet_url
}

output "yaml" {
  description = "Kubernetes deployment YAML"
  value       = module.api.yaml
}
output "role_arn" {
  description = "AWS role arn"
  value       = module.iam_eks_role.iam_role_arn
}

output "cloudbeat_ssh_cmd" {
  value = module.aws_ec2_with_agent.cloudbeat_ssh_cmd
  sensitive = true
}

output "ec2_private_key" {
  value = module.aws_ec2_with_agent.ec2_private_key
  sensitive = true
}
