# Gemini Assistant Context

This document provides context for AI assistants working on the Log Beacon project.

## Project Overview

Log Beacon is a log management system inspired by Humio. The primary goal is to build a performant and scalable backend service for ingesting, storing, and searching logs.

## Tech Stack

- **Language:** Go
- **Web Framework:** Gin (`github.com/gin-gonic/gin`)
- **Message Queue:** NATS (`github.com/nats-io/nats.go`)
- **Backend Focus:** The current development focus is purely on the backend. No frontend work is planned at this stage.

## High-Level Architecture

- **API Handler (`api` service):** A Go service using Gin that exposes the `/api/v1/ingest` endpoint. It validates incoming logs and publishes them to a NATS subject.
- **Message Queue (`nats` service):** A NATS server with JetStream enabled, acting as a durable buffer.
- **Consumer (`consumer` service):** A separate Go service that subscribes to the NATS subject, consumes the logs, and will eventually be responsible for writing them to persistent storage.
- **Storage (Planned):**
  - **Index (Metadata):** An embedded key-value store like BadgerDB or BoltDB.
  - **Log Data (Chunks):** A local filesystem or an object store (like MinIO).

## Current Status

- The project is managed via `docker-compose` with three services: `api`, `consumer`, and `nats`.
- Base images are pinned to specific versions for reproducibility.
- The `nats` service uses a persistent, host-bound volume for JetStream storage (`/tmp/log-beacon/nats-data`).
- The `api` service programmatically ensures the `LOGS` stream exists on startup.
- The Go application is structured into multiple binaries (`cmd/api`, `cmd/consumer`) and internal packages (`internal/server`, `internal/consumer`, `internal/model`, `internal/queue`).
- The `api` service successfully publishes logs to the `log.events` NATS subject.
- The `consumer` service subscribes to `log.events` and currently prints the consumed logs to the console, confirming the end-to-end pipeline is functional.
- The consumer logic has been refactored into a modular `internal/consumer` package.
