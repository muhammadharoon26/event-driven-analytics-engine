# -----------------------------------------------------------------------------
# Root Module — Event-Driven Analytics Engine
# Composes the Neon (PostgreSQL) and Upstash (Kafka) child modules to
# provision the full infrastructure stack for the analytics pipeline.
# -----------------------------------------------------------------------------

terraform {
  required_version = ">= 1.5"

  # Uncomment the backend block below to enable remote state storage.
  # backend "s3" {
  #   bucket = "your-terraform-state-bucket"
  #   key    = "event-driven-analytics/terraform.tfstate"
  #   region = "us-east-1"
  # }
}

# --- Neon PostgreSQL Module ---
# Provisions a Neon project, branch, and database for storing
# processed analytics data and serving queries.

module "neon" {
  source = "./neon"

  neon_api_key = var.neon_api_key
}

# --- Upstash Kafka Module ---
# Provisions an Upstash Kafka cluster and topic for ingesting
# real-time user events into the analytics pipeline.

module "upstash" {
  source = "./upstash"

  upstash_email   = var.upstash_email
  upstash_api_key = var.upstash_api_key
}
