terraform {
  required_providers {
    neon = {
      source  = "kislerdm/neon"
      version = ">= 0.1.2"
    }
  }
}

provider "neon" {
  api_key = var.neon_api_key
}

resource "neon_project" "analytics_project" {
  name                      = "event-driven-analytics"
  history_retention_seconds = 86400
}

resource "neon_branch" "main" {
  project_id = neon_project.analytics_project.id
  name       = "main"
}

resource "neon_database" "analytics_db" {
  project_id = neon_project.analytics_project.id
  branch_id  = neon_branch.main.id
  name       = "analytics"
  owner_name = "postgres"
}
