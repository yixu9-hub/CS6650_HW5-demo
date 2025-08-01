output "ecr_repository_url" {
  description = "URL of the ECR repo holding the image"
  value       = module.ecr.repository_url
}

output "ecs_cluster_name" {
  description = "Name of the created ECS cluster"
  value       = module.ecs.cluster_name
}

output "ecs_service_name" {
  description = "Name of the running ECS service"
  value       = module.ecs.service_name
}