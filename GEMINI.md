# Gemini Assistant Context

This document provides context for AI assistants working on the Log Beacon project.

## Project Overview

Log Beacon is a log management system inspired by Humio. It uses a hot/cold storage architecture to provide both fast, real-time search and durable, long-term storage.

## Tech Stack

- **Backend:** Go, Gin, NATS, MinIO, Bleve, BadgerDB
- **Frontend:** React, TypeScript, Vite, Bootstrap
- **Orchestration:** Docker Compose, Makefile

## High-Level Architecture

- **Frontend (`frontend` service):** A React-based single-page application that provides the user interface for searching logs. It communicates with the `api` service.
- **API Handler (`api` service):** A Go service using Gin that exposes `/api/v1/ingest` and `/api/v1/search`. It publishes logs to NATS and proxies search requests to the `hot-storage` service.
- **Message Queue (`nats` service):** A NATS server with JetStream enabled, acting as a durable buffer.
- **Cold Storage (`archiver` service):** A modular Go consumer that subscribes to NATS and writes all logs to a MinIO bucket for long-term archival.
- **Hot Storage (`hot-storage` service):** A modular Go consumer that subscribes to NATS and uses Bleve (full-text index) and BadgerDB (key-value store) to provide a fast, searchable index of recent logs. It also exposes an internal search API on port 8081.
- **Object Storage (`minio` service):** A MinIO server that stores the compressed log data from the `archiver`.

## Current Status

- The project is fully managed via a `Makefile` which automates directory creation, Docker Compose commands, and testing.
- The environment consists of six services: `nats`, `minio`, `api`, `archiver`, `hot-storage`, and `frontend`.
- Both the `archiver` and `hot-storage` services are implemented with a modular internal structure.
- The `frontend` service provides a functional and styled UI with pagination and dark mode for searching logs.
- Unit tests are integrated into the `make test` command.
- Persistent data volumes for all stateful services are managed via `docker-compose` and created on the host in `/tmp/log-beacon`.
