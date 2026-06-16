# -----------------------------------------------------------------------------
# Root Module Outputs
# Exposes key outputs from both the Neon and Upstash child modules.
# -----------------------------------------------------------------------------

# --- Neon (PostgreSQL) Outputs ---

output "neon_database_host" {
  description = "Connection endpoint for the Neon PostgreSQL database"
  value       = module.neon.database_host
  sensitive   = true
}

output "neon_database_name" {
  description = "Name of the provisioned analytics database"
  value       = module.neon.database_name
}

output "neon_project_id" {
  description = "Neon project ID"
  value       = module.neon.project_id
}

output "neon_branch_id" {
  description = "Neon branch ID for the main branch"
  value       = module.neon.branch_id
}

# --- Upstash (Kafka) Outputs ---

output "upstash_kafka_endpoint" {
  description = "Upstash Kafka cluster connection endpoint"
  value       = module.upstash.kafka_endpoint
  sensitive   = true
}

output "upstash_kafka_topic_name" {
  description = "Name of the Kafka topic for user events"
  value       = module.upstash.kafka_topic_name
}

output "upstash_cluster_id" {
  description = "Upstash Kafka cluster ID"
  value       = module.upstash.cluster_id
}
