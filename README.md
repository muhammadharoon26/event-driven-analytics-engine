# Event-Driven Analytics Engine

A decoupled, high-throughput microservices application demonstrating real-time data ingestion, message queuing, and persistent storage. Built with Go, Python, React, and Redpanda (Kafka).

![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)

## 🏗️ Architecture Stack

This project uses an event-driven pattern designed to decouple the high-speed API layer from the slower database storage layer.

- **Frontend:** React + Vite (Simulates high-frequency user events with optimistic UI and active DB polling)
- **Ingestion API:** Go + Gin (Accepts events instantaneously and pushes them to the queue)
- **Message Broker:** Redpanda / Kafka (Stores the event stream reliably to prevent data loss)
- **Data Processor:** Python + SQLAlchemy (A background worker that consumes events and writes to DB)
- **Database:** PostgreSQL (The persistent source-of-truth)

## 🚀 Quick Start (Local Development)

The easiest way to get the entire architecture running is to use the provided PowerShell script. It automatically uses Docker for the infrastructure (Redpanda, DB, Python Worker) and opens the Frontend/Backend in separate terminal windows for easy debugging.

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

1. Spin up **Postgres**, **Redpanda**, and the **Python Data Processor** in Docker backgrounds.
2. Run an init-container to automatically create the `user-events` Kafka topic.
3. Open a new window running the **Go Ingestion API** on `http://localhost:8080`.
4. Open a new window running the **React Frontend** on `http://localhost:5173`.

> **Note on Ports:** The local Postgres Docker container binds to host port `5433` to prevent conflicts with native Windows Postgres installations.

## 💾 The Real-Time Data Flow

1. You click an action in the React app (e.g., _Scan Product_).
2. The React app fires a `POST /api/v1/events` to the Go Backend.
3. The Go backend pushes the event into Redpanda and _immediately_ replies `{"status": "queued"}`. This allows the API to respond in `< 10ms`.
4. The React app begins polling the Go API (`GET /api/v1/events/status/:user_id`) every second.
5. In the background, the Python Data Processor pulls the event from Redpanda and writes it to Postgres.
6. The next time the React app polls the Go API, Go sees the database entry and returns `{"status": "persisted"}`, updating the UI with a green checkmark.

## 🛠️ Modifying the Services

### 1. Ingestion API (Go)

Located in `/ingestion-api`. If you add new dependencies, ensure you update the vendor folder:

```bash
cd ingestion-api
go mod tidy
go mod vendor
```

### 2. Frontend (React)

Located in `/frontend`. Uses Tailwind CSS and `framer-motion` for animations.

```bash
cd frontend
npm install
npm run dev
```

### 3. Data Processor (Python)

Located in `/data-processor`. It connects using `confluent-kafka` and writes to the DB using `SQLAlchemy`.
If you update this script, rebuild the docker container via:

```bash
docker-compose up -d --build data-processor
```

## 🧹 Teardown

To shut down the background Docker infrastructure:

```bash
docker-compose down -v
```

_(The `-v` flag removes the Database volumes so you start fresh next time)._
