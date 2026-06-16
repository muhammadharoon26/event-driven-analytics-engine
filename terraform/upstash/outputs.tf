output "kafka_endpoint" {
  description = "The Upstash Kafka cluster connection endpoint"
  value       = upstash_kafka_cluster.analytics_cluster.rest_endpoint
  sensitive   = true
}

output "kafka_topic_name" {
  description = "The name of the Kafka topic for user events"
  value       = upstash_kafka_topic.user_events.topic_name
}

output "cluster_id" {
  description = "The Upstash Kafka cluster ID"
  value       = upstash_kafka_cluster.analytics_cluster.cluster_id
}
