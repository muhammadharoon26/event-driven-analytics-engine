# Event-Driven Analytics Engine: Product Profile

## Overview

The **Event-Driven Analytics Engine** is a high-throughput, decoupled microservices architecture designed to demonstrate real-time data ingestion, message queuing, and persistent storage. Built with modern, scalable technologies, this project serves as a foundational blueprint for handling massive streams of user actions (e.g., clicks, page views, product scans) asynchronously.

By decoupling the data ingestion layer from the database storage layer using a Kafka-compatible message broker, the system guarantees high availability, rapid response times, and zero data loss during traffic spikes.

---

## Core Objectives

- **Asynchronous Ingestion:** Ensure the API responds instantly (`< 10ms`) to client requests by offloading database writes to background workers.
- **Fault Tolerance:** Utilize message queues to buffer events. If the database goes offline, incoming events are safely held in the queue until service is restored.
- **Scalability:** Allow independent scaling of the API layer (to handle incoming traffic) and the processing layer (to handle database writes).
- **Real-time Observability:** Provide a frontend that accurately traces and visualizes the lifecycle of an event from ingestion to database persistence.

---

## System Architecture

The project follows a standard Event-Driven Microservices pattern, divided into four distinct components:

1. **Frontend (React + Vite)**
   - **Role:** Simulates a client application generating high-frequency user events.
   - **Features:** A dynamic UI with `framer-motion` animations, real-time metrics tracking, and an active polling mechanism that visually traces the journey of an event (`API` ➔ `Kafka` ➔ `Postgres`).

2. **Ingestion API (Go + Gin)**
   - **Role:** The high-performance gatekeeper.
   - **Features:** Exposes RESTful endpoints (`POST /api/v1/events`). It validates incoming JSON payloads and instantly pushes them to the message broker. It also features a real-time polling endpoint (`GET /api/v1/events/status/:user_id`) utilizing `pgxpool` to actively check the database for event confirmation.

3. **Message Broker (Redpanda / Kafka-compatible)**
   - **Role:** The resilient middle layer.
   - **Features:** Safely stores incoming events in the `user-events` topic. It acts as a shock absorber, decoupling the fast ingestion API from the slower database layer. Includes an initialization container to auto-create missing topics.

4. **Data Processor (Python + SQLAlchemy/Confluent-Kafka)**
   - **Role:** The background worker.
   - **Features:** Continuously tails the Redpanda topic, validates the incoming data streams, and performs batch writes to the PostgreSQL database. Designed to fail gracefully and utilize Dead Letter Queues (DLQ) for malformed data.

5. **Storage Container (PostgreSQL)**
   - **Role:** The source of truth.
   - **Features:** Relational database storing the sanitized analytics events inside the `analytics` database volume.

---

## Tech Stack

- **Frontend:** React, Vite, Tailwind CSS, Framer Motion, Axios, Lucide React
- **Backend API:** Go (1.25), Gin Web Framework, Segmentio Kafka-Go, PGX (Postgres Driver)
- **Message Broker:** Redpanda (Kafka v23.2.19)
- **Worker Service:** Python 3.x, Confluent-Kafka, Pydantic, SQLAlchemy
- **Database:** PostgreSQL 15 (Alpine)
- **DevOps / Orchestration:** Docker, Docker Compose, PowerShell

---

## Data Flow Lifecycle

1. **User Action:** A user clicks "Add to Cart" on the UI.
2. **REST POST:** The React frontend fires an event payload to the Go API.
3. **Queue Push:** Go API pushes the message to Redpanda and replies `202 Accepted {"status": "queued"}`.
4. **Client Polling:** React immediately begins 1-second interval polling against the Go status endpoint.
5. **Background Read:** The Python processor consumes the event from Redpanda.
6. **Persistence:** Python processor writes the event into PostgreSQL `analytics_events` table.
7. **Client Confirmation:** On the next React poll, the Go API detects the Postgres row and returns `200 OK {"status": "persisted"}`, updating the UI with a green `Postgres` checkmark.
