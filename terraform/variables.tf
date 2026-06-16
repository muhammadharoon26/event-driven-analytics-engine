# -----------------------------------------------------------------------------
# Root Module Variables
# All variables required by the Neon and Upstash child modules
# -----------------------------------------------------------------------------

# --- Neon (PostgreSQL) Variables ---

variable "neon_api_key" {
  description = "API key for Neon.tech to provision the PostgreSQL database"
  type        = string
  sensitive   = true
}

# --- Upstash (Kafka) Variables ---

variable "upstash_email" {
  description = "Email address associated with the Upstash account"
  type        = string
}

variable "upstash_api_key" {
  description = "API key for Upstash to provision the Kafka cluster"
  type        = string
  sensitive   = true
}
