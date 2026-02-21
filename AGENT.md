# Agent Instructions for Log Beacon

This document provides essential technical context for AI agents working on the Log Beacon repository.

## 1. Project Goal

The primary goal is to build a performant, scalable, and decoupled log management system inspired by Humio, featuring both fast real-time search (hot storage) and durable long-term archival (cold storage).

## 2. Core Technologies

- **Backend:** Go (Gin, NATS, Bleve, BadgerDB)
- **Frontend:** React, TypeScript, Vite, Tailwind CSS
- **Object Storage:** MinIO (S3-compatible)
- **Message Queue:** NATS with JetStream
- **Containerization:** Docker and Docker Compose
- **Orchestration:** Makefile

## 3. Architecture Overview

The system operates as a multi-service application orchestrated by `docker-compose.yml`.

1.  **`api` Service:**
    - Entry point for ingestion (`POST /api/v1/ingest`), search proxying, and WebSocket-based live tailing (`/api/v1/tail`).
2.  **`nats` Service:**
    - Durable message buffer using JetStream. Persistent data stored in `/tmp/log-beacon/nats-data`.
3.  **`hot-storage` Service:**
    - Consumer that indexes recent logs using Bleve and BadgerDB for fast searching. Exposes an internal search API on port 8081.
4.  **`archiver` Service:**
    - Consumer that writes all logs to MinIO for long-term archival.
5.  **`minio` Service:**
    - S3-compatible object storage for archived logs.
6.  **`frontendv2` Service:**
    - React-based single-page application for searching and viewing logs.

## 4. How to Build & Run

The entire development environment is managed via a `Makefile`.

- **To build and start all services:** `make` or `make up`
- **To stop all services:** `make down`
- **To follow logs:** `make logs`
- **To run unit tests:** `make test`
- **To clean persistent data:** `make clean`

## 5. Key Design Decisions

- **Hot/Cold Storage:** Separating "hot" (recent, indexed) and "cold" (old, archived) data allows for fast searches on recent logs while maintaining cost-effective long-term storage.
- **Microservices Decoupling:** Ingestion, indexing, and archival are separate services connected via NATS, allowing independent scaling and decoupling of concerns.
- **Search Refinement:** Supports structured queries (e.g., `service:auth AND level:error`) with automatic label rewriting (e.g., `service:auth` -> `labels.service:auth`).
- **Live Tail:** Powered by WebSockets in the `api` service, providing real-time log streaming directly to the `frontendv2` UI.
- **Persistence:** All stateful data is mapped to `/tmp/log-beacon` on the host machine for persistence across container restarts.
