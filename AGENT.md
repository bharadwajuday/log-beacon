# Agent Instructions for Log Beacon

This document provides essential technical context, architecture overview, and design decisions for AI agents working on the Log Beacon repository.

## 1. Project Goal & Overview
Log Beacon is a high-performance, Humio-inspired log ingestion and search platform built with Go and React. It uses a decoupled, microservices-oriented architecture with a hot/cold storage strategy to provide both fast, real-time search (hot storage) and durable, long-term storage (cold storage).

## 2. Core Technologies
- **Backend:** Go (Gin, NATS, Bleve, BadgerDB)
- **Frontend:** React, TypeScript, Vite, Tailwind CSS
- **Object Storage:** MinIO (S3-compatible)
- **Message Queue:** NATS with JetStream
- **Database:** Postgres (for authentication and system metadata)
- **Containerization:** Docker and Docker Compose
- **Orchestration:** Makefile

## 3. High-Level Architecture
The system operates as a multi-service application orchestrated by `docker-compose.yml`:
1. **`api` Service:** Entry point for ingestion (`POST /api/v1/ingest`), search proxying (`POST /api/v1/search`), and WebSocket-based live tailing (`/api/v1/tail`).
2. **`nats` Service:** Durable message buffer using JetStream. Persistent data is stored in `$HOME/log-beacon-data/nats-data`.
3. **`hot-storage` Service:** A modular Go consumer that subscribes to NATS and uses Bleve (full-text index) and BadgerDB (key-value store) to provide a fast, searchable index of recent logs. Exposes an internal search API on port `8081`.
4. **`archiver` Service:** A modular Go consumer that writes all logs to MinIO for long-term archival.
5. **`minio` Service:** S3-compatible object storage for archived logs.
6. **`postgres` Service:** Relational database for storing user credentials and authentication metadata.
7. **`frontendv2` Service:** React-based single-page application for searching and viewing logs.

## 4. Current Status & Key Features
- **Search Refinement:** The `hot-storage` service supports structured queries with `AND`/`OR` operators and automatic field rewriting (e.g., `service:auth` -> `labels.service:auth`).
- **Frontend Integration:** The `frontendv2` UI supports implicit log level filtering using the sidebar, which is combined with the main search query.
- **WebSocket Live Tail:** The `api` service exposes a WebSocket endpoint at `/api/v1/tail` for real-time log streaming, fully integrated into the `frontendv2` UI.
- **Authentication:** JWT-based authentication is implemented across the stack, protecting search, ingestion, and live tail endpoints.
- **Columnar Cold Storage**: The `archiver` service buffers logs in memory and writes compressed, Hive-partitioned Apache Parquet files (`logs/year=YYYY/month=MM/day=DD/hour=HH/*.parquet`) to MinIO.
- **Hot Storage Retention**: The `hot-storage` service runs a periodic background cleaner (`PurgeExpiredLogs`) purging logs older than `HOT_STORAGE_RETENTION` (default 24h) from BadgerDB and Bleve.
- **Federated Search**: The `api` service provides transparent federated querying across recent hot storage and historical Parquet cold archives.
- **Kubernetes Deployment**: Complete production Kubernetes manifests and Kustomize specs are available in `k8s/`.
- **Testing:** Unit tests are integrated into the `make test` command.

## 5. Key Design Decisions
- **Hot/Cold Storage:** Separating "hot" (recent, indexed) and "cold" (old, archived) data allows for fast searches on recent logs while maintaining cost-effective long-term storage.
- **Microservices Decoupling:** Ingestion, indexing, and archival are separate services connected via NATS, allowing independent scaling and decoupling of concerns.
- **Dynamic Log Schema:** The log entry structure uses custom JSON unmarshaling to map arbitrary top-level fields into a dynamic `Labels` map (`internal/model/log.go`), ensuring extensibility without mutating the shared log type.
- **Persistence:** All stateful data is mapped to `$HOME/log-beacon-data` on the host machine for persistence across container restarts.

## 6. How to Build & Run
The entire development environment is managed via a `Makefile`.
- **Build and start all services:** `make` or `make up`
- **Stop all services:** `make down`
- **Follow logs:** `make logs`
- **Run unit tests:** `make test`
- **Clean persistent data:** `make clean`
