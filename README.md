# Event-Driven Analytics Engine

A production-grade, **event-driven microservices platform** demonstrating real-time data ingestion, distributed message queuing, persistent storage, infrastructure-as-code, and end-to-end **distributed tracing**.

Measured on a single laptop running the whole stack: **~13,800 events/sec** accepted by the ingestion API at a **p50 of 6.6ms**, and **~500–1,100 events/sec** carried all the way through to Postgres. See [Reproducing the Performance Numbers](#-reproducing-the-performance-numbers) to run the benchmark yourself.

![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)
![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)
![React](https://img.shields.io/badge/React-19-61DAFB?logo=react&logoColor=white)
![Python](https://img.shields.io/badge/Python-3.11-3776AB?logo=python&logoColor=white)
![Terraform](https://img.shields.io/badge/Terraform-IaC-7B42BC?logo=terraform&logoColor=white)

---

## 🏗️ Architecture

```
┌──────────────┐       ┌──────────────────┐       ┌────────────────┐       ┌──────────────┐
│  React + Vite│──────▶│  Go (Gin) API    │──────▶│  Redpanda      │──────▶│  Python      │
│  Dashboard   │  HTTP │  Ingestion Layer │ Kafka │  (Kafka-compat)│  Msg  │  Processor   │
│              │◀──poll│  <10ms response  │       │  Message Broker│       │  SQLAlchemy  │
└──────────────┘       └──────────────────┘       └────────────────┘       └──────┬───────┘
                              │                                                    │
                              │ traces                                    writes   │
                              ▼                                                    ▼
                       ┌──────────────────┐                              ┌──────────────┐
                       │  OpenTelemetry   │                              │  PostgreSQL  │
                       │  Collector       │                              │  Database    │
                       └────────┬─────────┘                              └──────────────┘
                                │
                                ▼
                       ┌──────────────────┐       ┌──────────────┐
                       │  Grafana Tempo   │──────▶│   Grafana    │
                       │  Trace Backend   │       │  Dashboards  │
                       └──────────────────┘       └──────────────┘
```

### Technology Stack

| Layer | Technology | Purpose |
|---|---|---|
| **Frontend** | React 19 + Vite | Real-time dashboard with optimistic UI and live event polling |
| **API Gateway** | Go (Gin) | High-performance HTTP ingestion, measured p50 6.6ms under load |
| **Message Broker** | Redpanda (Kafka-compatible) | Distributed event streaming, absorbs bursts the database cannot |
| **Data Pipeline** | Python + SQLAlchemy + Pydantic | Event validation, transformation, and persistence |
| **Database** | PostgreSQL 15 | ACID-compliant persistent storage for analytics events |
| **IaC** | Terraform | Provisioning of cloud infrastructure (Neon PostgreSQL + Upstash Kafka) |
| **GitOps** | ArgoCD | Declarative, automated Kubernetes deployment from Git |
| **Observability** | OpenTelemetry + Grafana Tempo + Grafana | End-to-end distributed tracing across all microservices |
| **Containers** | Docker + Docker Compose | Local development and CI/CD pipeline |
| **Orchestration** | Kubernetes | Production deployment manifests with health checks and resource limits |
| **CI/CD** | GitHub Actions | Automated build, test, and Docker image publishing |

---

## 🚀 Quick Start (Local Development)

### Prerequisites

- [Docker & Docker Compose](https://docs.docker.com/get-docker/)
- [Go 1.25+](https://golang.org/doc/install)
- [Node.js 18+](https://nodejs.org/en/download/)

### Running the Project

```powershell
# Run the startup script
.\start.ps1
```

The script will:

1. Spin up **Postgres**, **Redpanda**, **Python Data Processor**, **OTEL Collector**, **Tempo**, and **Grafana** in Docker.
2. Run an init-container to automatically create the `user-events` Kafka topic.
3. Open a new terminal running the **Go Ingestion API** on `http://localhost:8080`.
4. Open a new terminal running the **React Frontend** on `http://localhost:5173`.

### Access Points

| Service | URL | Description |
|---|---|---|
| React Dashboard | http://localhost:5173 | Interactive event firing UI |
| Go API | http://localhost:8080 | REST API with health check at `/health` |
| Grafana | http://localhost:3000 | Distributed tracing dashboards |
| Redpanda | localhost:9092 | Kafka-compatible broker |
| PostgreSQL | localhost:5433 | Analytics database |

> **Note on Ports:** The local Postgres Docker container binds to host port `5433` to prevent conflicts with native Windows Postgres installations.

---

## 💾 The Real-Time Data Flow

1. You click an action in the **React** app (e.g., _Scan Product_).
2. The React app fires a `POST /api/v1/events` to the **Go (Gin)** backend.
3. The Go backend pushes the event into **Redpanda** (Kafka) and _immediately_ replies `{"status": "queued"}`. The **OpenTelemetry** middleware records the full request trace and response time as a span — consistently `< 10ms`.
4. The React app begins polling the Go API (`GET /api/v1/events/status/:user_id`) every second.
5. In the background, the **Python** Data Processor pulls the event from Redpanda, validates it with **Pydantic**, buffers it, and commits it to **PostgreSQL** via **SQLAlchemy** together with the rest of its batch — all under traced spans. Kafka offsets are committed only after that write succeeds, which is what makes delivery at-least-once.
6. The next poll returns `{"status": "persisted"}`, updating the UI with a green checkmark.
7. The full distributed trace (Go → Kafka → Python → Postgres) is viewable in **Grafana** via **Tempo**.

---

## 📊 Distributed Tracing (OpenTelemetry + Grafana)

Both microservices are instrumented with **OpenTelemetry**:

- **Go API**: Every HTTP request creates a span with `http.method`, `http.route`, `http.status_code`, and `response_time_ms` attributes. A latency middleware adds the `X-Response-Time` header, which the dashboard reads back to show the real server-measured time.
- **Python Processor**: Each message creates `consume_event` and `validate_event` spans carrying `event.user_id` and `event.action`. Because rows are written in batches, the database write is a separate `persist_batch` span with a `batch.size` attribute rather than one span per row.

Traces are exported via OTLP gRPC to an **OpenTelemetry Collector**, which forwards to **Grafana Tempo**. **Grafana** provides a pre-configured dashboard for exploring traces.

```
Go/Python → OTLP (gRPC:4317) → OTEL Collector → Tempo → Grafana
```

---

## ☁️ Infrastructure as Code (Terraform)

Cloud infrastructure is provisioned with **Terraform** using two modules:

```
terraform/
├── main.tf              # Root module composing Neon + Upstash
├── variables.tf         # Root-level variables
├── outputs.tf           # Aggregated outputs
├── neon/                # PostgreSQL on Neon.tech
│   ├── main.tf
│   ├── variables.tf
│   └── outputs.tf
└── upstash/             # Kafka on Upstash
    ├── main.tf
    ├── variables.tf
    └── outputs.tf
```

```bash
cd terraform
terraform init
terraform plan -var="neon_api_key=..." -var="upstash_api_key=..." -var="upstash_email=..."
terraform apply
```

---

## 🔄 GitOps with ArgoCD

The `k8s/` directory contains Kubernetes manifests managed by **ArgoCD**:

```
k8s/
├── argocd-app.yaml       # ArgoCD Application (auto-sync + self-heal)
├── configmap.yaml        # Environment configuration + secrets
├── ingestion-api.yaml    # Go API Deployment + Service (with health checks)
├── data-processor.yaml   # Python Processor Deployment (with resource limits)
├── otel-collector.yaml   # OTEL Collector Deployment + ConfigMap
└── grafana.yaml          # Grafana + Tempo Deployments
```

ArgoCD watches this repo and automatically syncs changes:
- **Automated sync** with self-healing enabled
- **Pruning** of orphaned resources
- **Namespace auto-creation** for `analytics-engine`

---

## 🛠️ Modifying the Services

### 1. Ingestion API (Go + Gin)

Located in `/ingestion-api`. Uses Gin for routing, `kafka-go` for Kafka, and OpenTelemetry for tracing.

```bash
cd ingestion-api
go mod tidy
go mod vendor
go run main.go
```

### 2. Frontend (React + Vite)

Located in `/frontend`. Uses Tailwind CSS and `framer-motion` for animations.

```bash
cd frontend
npm install
npm run dev
```

### 3. Data Processor (Python)

Located in `/data-processor`. Uses `confluent-kafka` for Kafka consumption, `SQLAlchemy` for ORM, and OpenTelemetry for tracing.

```bash
cd data-processor
pip install -r requirements.txt
python main.py
```

---

## 🧪 Testing

```bash
# Go unit tests
cd ingestion-api && go test ./...

# Python unit tests
cd data-processor && python -m pytest test_main.py
```

---

## 📈 Reproducing the Performance Numbers

The throughput and latency figures above are not estimates — `ingestion-api/cmd/loadtest`
measures them against a running stack.

```bash
# Start everything first (.\start.ps1), then:
cd ingestion-api
go run ./cmd/loadtest -n 20000 -c 100
```

Measured on a Windows 11 laptop with the entire stack (Redpanda, Postgres,
Grafana, Tempo, OTEL Collector, the Python worker, the Go API and the load
generator) sharing one machine:

| Metric | Result |
|---|---|
| Ingestion throughput | **~13,800 events/sec** accepted onto the Kafka topic |
| API latency, 100 concurrent | p50 **6.6ms**, p95 **15.7ms**, p99 **23.7ms** |
| API latency, 10 concurrent | p50 **1.7ms**, p95 **8.8ms** |
| End-to-end pipeline | **~500–1,100 events/sec** persisted to Postgres |

The API and the worker are deliberately different numbers. The API only has to
put the event on the topic, so it is fast; the worker has to validate every
event and commit it to Postgres, so it is the narrower part of the pipe. That
gap is the entire point of putting a queue between them — a traffic spike
lands in Redpanda instead of knocking the database over.

Confirm the worker drained what the API accepted:

```bash
docker exec postgres psql -U postgres -d analytics -c "SELECT count(*) FROM analytics_events;"
```

---

## 🧹 Teardown

To shut down the background Docker infrastructure:

```bash
docker-compose down -v
```

_(The `-v` flag removes the database volumes so you start fresh next time)._

---

## 📁 Project Structure

```
event-driven-analytics-engine/
├── .github/workflows/      # CI/CD pipeline (build, test, Docker push)
├── frontend/                # React + Vite dashboard
├── ingestion-api/           # Go (Gin) high-performance API
│   ├── kafka/               # Kafka producer package
│   ├── telemetry/           # OpenTelemetry tracer + middleware
│   ├── cmd/loadtest/        # Benchmark that produces the numbers above
│   ├── main.go              # API entrypoint
│   └── Dockerfile
├── data-processor/          # Python event consumer + DB writer
│   ├── database/            # SQLAlchemy models + session
│   ├── telemetry/           # OpenTelemetry tracer
│   ├── main.py              # Consumer entrypoint
│   └── Dockerfile
├── k8s/                     # Kubernetes manifests + ArgoCD
├── terraform/               # Infrastructure as Code (Neon + Upstash)
├── observability/           # OTEL Collector, Tempo, Grafana configs
├── docker-compose.yml       # Full local development stack
└── start.ps1                # One-click startup script
```
