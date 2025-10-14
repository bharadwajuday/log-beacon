# Gemini Assistant Context

This document provides context for AI assistants working on the Log Beacon project.

## Project Overview

Log Beacon is a log management system inspired by Humio. The primary goal is to build a performant and scalable backend service for ingesting, storing, and searching logs.

## Tech Stack

- **Language:** Go
- **Web Framework:** Gin (`github.com/gin-gonic/gin`)
- **Backend Focus:** The current development focus is purely on the backend. No frontend work is planned at this stage.

## High-Level Architecture

- **API Handler:** A Go service using Gin that exposes the `/api/v1/ingest` endpoint. Its sole responsibility is to receive log data, validate it, and publish it to a NATS message queue.
- **Message Queue:** A NATS server (with JetStream enabled) acts as a durable buffer between the API handler and downstream processing services.
- **Consumer (Planned):** A separate Go service that will consume logs from the NATS queue, process them, and write them to a persistent storage layer.
- **Storage (Planned):**
  - **Index (Metadata):** An embedded key-value store like BadgerDB or BoltDB.
  - **Log Data (Chunks):** A local filesystem or an object store (like MinIO).

## Current Status

- The project is managed via `docker-compose` with pinned image versions (`nats:2.10.14-alpine`, `alpine:3.19`).
- The `nats` service is configured with a persistent, host-bound volume for JetStream storage located at `/tmp/log-beacon/nats-data`.
- The `api` service is defined and containerized.
- On startup, the `api` service programmatically ensures a durable NATS stream named `LOGS` (for the `log.events` subject) exists.
- The Go application is structured into `main`, `internal/server`, `internal/model`, and `internal/queue` packages.
- The `handleIngest` endpoint currently parses incoming logs but does not yet publish them to NATS.
