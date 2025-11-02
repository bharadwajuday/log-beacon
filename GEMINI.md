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

- **API Handler (`api` service):** A Go service using Gin that exposes the `/api/v1/ingest` endpoint. It validates incoming logs and publishes them to a NATS subject.
- **Message Queue (`nats` service):** A NATS server with JetStream enabled, acting as a durable buffer.
- **Consumer (`consumer` service):** A separate Go service that subscribes to the NATS subject, consumes the logs, compresses them, and writes them as gzipped objects to a MinIO bucket.
- **Storage (`minio` service):**
  - **Log Data (Chunks):** A MinIO server stores the compressed log data.
  - **Index (Metadata - Planned):** An embedded key-value store like BadgerDB or BoltDB is planned for the consumer.

## Current Status

- The project is managed via `docker-compose` with four services: `api`, `consumer`, `nats`, and `minio`.
- All services use pinned image versions and persistent, host-bound volumes for data.
- The `api` service programmatically ensures the `LOGS` NATS stream exists on startup.
- The `consumer` service programmatically ensures the `logs` MinIO bucket exists on startup.
- A modular `internal/storage` package with an `ObjectStorage` interface has been implemented.
- The end-to-end pipeline is fully functional: logs sent to the API are published to NATS, consumed, gzipped, and stored as objects in MinIO.
