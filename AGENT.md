# Agent Instructions for Log Beacon

This document provides essential technical context for AI agents working on the Log Beacon repository.

## 1. Project Goal

The primary goal is to build a performant, scalable, and decoupled backend service for ingesting, storing, and searching application logs.

## 2. Core Technologies

- **Language:** Go
- **API:** Gin (`github.com/gin-gonic/gin`)
- **Message Queue:** NATS with JetStream (`github.com/nats-io/nats.go`)
- **Containerization:** Docker and Docker Compose

## 3. Architecture Overview

The system operates as a multi-service application orchestrated by `docker-compose.yml`.

1. **`api` Service:**
    - A Go application that runs a Gin web server.
    - Exposes the `POST /api/v1/ingest` endpoint.
    - **Responsibility:** Receives log data, validates it, and publishes it as a message to the `log.events` subject on the NATS server.

2. **`nats` Service:**
    - Runs a NATS server with JetStream enabled for persistence.
    - **Responsibility:** Acts as a durable, resilient buffer for all incoming log messages. Data is stored on a host-bound volume in `/tmp/log-beacon/nats-data`.

3. **`consumer` Service:**
    - A separate Go application that acts as a background worker.
    - **Responsibility:** Subscribes to the `log.events` subject, consumes the log messages, and will be responsible for processing and persisting them to long-term storage.

## 4. How to Build & Run

The entire development environment is managed by Docker Compose.

- **To build and run all services:**

    ```bash
    docker-compose up --build
    ```

- **To stop all services:**

    ```bash
    docker-compose down
    ```

## 5. Key Design Decisions

- **Decoupling with NATS:** The `api` and `consumer` services are intentionally decoupled. This allows the ingestion endpoint to be extremely fast and responsive, while the heavy lifting of processing and storage can be scaled independently. The use of a persistent message queue prevents data loss during traffic spikes or consumer downtime.
- **Modular Go Packages:** The Go code is organized into distinct `internal` packages (`server`, `consumer`, `queue`, `model`). This promotes code reuse and maintainability, and makes it easier to understand the different logical components of the system.
- **Multi-Binary Build:** The `Dockerfile` builds two separate Go binaries (`api` and `consumer`) into a single, shared final image. This is efficient and simplifies the deployment configuration in `docker-compose.yml`.
