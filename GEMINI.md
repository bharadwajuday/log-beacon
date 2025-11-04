# Gemini Assistant Context

This document provides context for AI assistants working on the Log Beacon project.

## Project Overview

Log Beacon is a log management system inspired by Humio. The primary goal is to build a performant and scalable backend service for ingesting, storing, and searching logs.

## Tech Stack

- **Language:** Go
- **Web Framework:** Gin (`github.com/gin-gonic/gin`)
- **Message Queue:** NATS (`github.com/nats-io/nats.go`)
- **Object Storage:** MinIO (`github.com/minio/minio-go/v7`)
- **Backend Focus:** The current development focus is purely on the backend. No frontend work is planned at this stage.

## High-Level Architecture

Log Beacon now uses a hot/cold storage architecture to provide both fast, real-time search and durable, long-term storage.

-   **API Handler (`api` service):** A Go service using Gin that exposes the `/api/v1/ingest` endpoint. It validates incoming logs and publishes them to a NATS subject.
-   **Message Queue (`nats` service):** A NATS server with JetStream enabled, acting as a durable buffer.
-   **Cold Storage (`archiver` service):** A Go consumer that subscribes to the NATS subject, compresses logs, and writes them as gzipped objects to a MinIO bucket for long-term archival.
-   **Hot Storage (`hot-storage` service - Planned):** A new Go consumer that will also subscribe to the NATS subject. It will use Bleve (full-text index) and BadgerDB (key-value store) to provide a fast, searchable index of recent logs.
-   **Object Storage (`minio` service):**
    -   **Log Data (Chunks):** A MinIO server stores the compressed log data from the `archiver`.

## Current Status

-   The project is managed via a `Makefile` which automates directory creation and `docker-compose` commands.
-   The environment consists of five services: `nats`, `minio`, `api`, `archiver`, and `hot-storage`.
-   The `archiver` service (previously `consumer`) is fully functional and writes logs to MinIO.
-   A placeholder `hot-storage` service has been created and is running in a container.
-   Persistent data volumes for all stateful services are managed via `docker-compose` and created on the host in `/tmp/log-beacon`.
