terraform {
  required_providers {
    upstash = {
      source  = "upstash/upstash"
      version = "~> 1.5"
    }
  }
}

provider "upstash" {
  email   = var.upstash_email
  api_key = var.upstash_api_key
}

resource "upstash_kafka_cluster" "analytics_cluster" {
  cluster_name = "analytics-events"
  region       = "us-east-1"
  multizone    = false
}

resource "upstash_kafka_topic" "user_events" {
  topic_name     = "user-events"
  partitions     = 1
  retention_time = 86400000
  retention_size = 1073741824
  cluster_id     = upstash_kafka_cluster.analytics_cluster.cluster_id
}
