output "database_host" {
  description = "The connection endpoint for the Neon PostgreSQL database"
  value       = neon_project.analytics_project.database_host
  sensitive   = true
}

output "database_name" {
  description = "The name of the provisioned analytics database"
  value       = neon_database.analytics_db.name
}

output "project_id" {
  description = "The Neon project ID"
  value       = neon_project.analytics_project.id
}

output "branch_id" {
  description = "The Neon branch ID for the main branch"
  value       = neon_branch.main.id
}
