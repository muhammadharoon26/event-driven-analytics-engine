# **Technical Design Document (TDD): Event-Driven Analytics & Notification Engine**

## **1\. Executive Summary**

**Objective:** To design and implement a highly scalable, zero-cost, enterprise-grade event-driven architecture capable of ingesting, queueing, and processing high volumes of asynchronous user actions.

**Use Case:** An analytics and notification engine that safely captures user interactions (e.g., product scans, button clicks) under heavy load without dropping requests or overwhelming the storage layer.

**Target Audience:** Technical recruiters assessing systems design, microservices architecture, and modern DevOps/GitOps proficiency.

## **2\. Architecture Overview**

The system employs a decoupled, 3-tier microservices architecture utilizing a message broker to ensure high availability and fault tolerance.

### **The 3-Tier Flow:**

1. **Ingestion (Tier 1):** A client sends a JSON payload to the API Gateway. The API validates the payload, pushes it to a message broker, and immediately returns a 202 Accepted response.  
2. **Message Queueing:** The message broker securely holds the event, buffering the system against traffic spikes.  
3. **Processing & Storage (Tiers 2 & 3):** A background worker continuously polls the broker, pulls the event, executes business logic (data enrichment/deduplication), and persists the finalized data to a relational database.

## **3\. Technology Stack & Infrastructure**

This project strictly utilizes scalable free-tier SaaS and cloud providers to simulate an enterprise environment without financial overhead.

| Component | Technology | Provider / Strategy |
| :---- | :---- | :---- |
| **Cloud Environment** | Kubernetes | Oracle Cloud Always Free (up to 4 ARM compute instances) or Local Minikube. |
| **Infrastructure as Code** | Terraform | Local execution or Terraform Cloud (Free tier for up to 500 resources). |
| **API Gateway (Ingestion)** | Go (Gin Framework) | High-performance, low-footprint REST/gRPC API. |
| **Message Broker** | Apache Kafka | Upstash (Serverless Kafka, 10,000 messages/day free). |
| **Data Processor** | Python (FastAPI) | Asynchronous consumer for data formatting and logic execution. |
| **Database** | PostgreSQL | Neon.tech (Serverless Postgres free tier). |
| **CI/CD Pipeline** | GitHub Actions | Automated testing and Docker image builds (2,000 free minutes/month). |
| **GitOps Deployment** | ArgoCD | Automated, pull-based synchronization of Kubernetes manifests. |
| **Observability** | Prometheus, OpenTelemetry | Grafana Cloud (Forever Free tier: 10k series metrics, 50GB logs/traces). |

## **4\. Component Level Design**

### **4.1. The Ingestion API (Go)**

* **Responsibility:** Handle high-throughput incoming traffic with minimal latency.  
* **Endpoint:** POST /api/v1/events  
* **Payload Example:** \`\`\`json  
  {  
  "user\_id": "12345",  
  "action": "scan\_started",  
  "timestamp": "2026-03-09T21:19:19Z"  
  }  
* **Behavior:** Performs basic schema validation. On success, publishes the payload to the user-events Kafka topic and closes the HTTP connection.

### **4.2. The Message Broker (Kafka via Upstash)**

* **Responsibility:** Act as a shock absorber. If the database goes offline or traffic spikes exponentially, data is safely retained in the queue.  
* **Topic Configuration:** Configured with appropriate retention policies to ensure data isn't lost before the Python processor can consume it.

### **4.3. The Data Processor (Python)**

* **Responsibility:** Consume messages, execute business logic, and manage data persistence.  
* **Behavior:** Runs as a continuous background process. It implements robust error handling (e.g., dead-letter queues for malformed messages) and writes the cleaned, structured data into PostgreSQL.

## **5\. DevOps & CI/CD Lifecycle**

The deployment lifecycle is heavily automated to showcase modern infrastructure management.

1. **Provisioning (IaC):** Terraform scripts define and provision the Neon PostgreSQL database and Upstash Kafka cluster.  
2. **Continuous Integration:** Commits to the main branch trigger GitHub Actions. The pipeline executes go test and pytest. Upon success, it builds Docker images and pushes them to Docker Hub.  
3. **Continuous Deployment (GitOps):** The GitHub Action updates a Kubernetes deployment manifest with the new image tag. ArgoCD detects the configuration drift and automatically syncs the live Kubernetes cluster to match the repository.  
4. **Monitoring:** Both Go and Python microservices are instrumented with OpenTelemetry. Metrics and traces are scraped by Grafana Cloud, providing real-time dashboards for API latency, Kafka queue depth, and database query performance.

## **6\. Key Selling Points for Interviews**

* **Resilience via Decoupling:** Proves an understanding that synchronous database writes are a bottleneck at scale.  
* **Modern Delivery:** Highlights the transition from traditional push-based deployments to GitOps (ArgoCD).  
* **Production Readiness:** The inclusion of OpenTelemetry and Grafana demonstrates the ability to maintain and debug complex systems after deployment.