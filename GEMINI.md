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

Log Beacon uses a hot/cold storage architecture to provide both fast, real-time search and durable, long-term storage.

- **Frontend (`frontend` service):** A React-based single-page application that provides the user interface for searching logs. It communicates with the `api` service.
- **API Handler (`api` service):** A Go service using Gin that exposes the `/api/v1/ingest` and `/api/v1/search` endpoints. It validates incoming logs, publishes them to NATS, and proxies search requests to the `hot-storage` service.
- **Message Queue (`nats` service):** A NATS server with JetStream enabled, acting as a durable buffer.
- **Cold Storage (`archiver` service):** A Go consumer that subscribes to the NATS subject and writes logs to a MinIO bucket for long-term archival.
- **Hot Storage (`hot-storage` service):** A Go consumer that subscribes to the NATS subject and uses Bleve and BadgerDB to provide a fast, searchable index of recent logs. It also exposes an internal search API.
- **Object Storage (`minio` service):** A MinIO server that stores the compressed log data from the `archiver`.

## Current Status

- The project is managed via a `Makefile` which automates directory creation and `docker-compose` commands.
- The environment consists of six services: `nats`, `minio`, `api`, `archiver`, `hot-storage`, and `frontend`.
- The `archiver` service writes logs to MinIO.
- The `hot-storage` service indexes logs and serves search queries via an internal API.
- The `frontend` service provides a functional and styled UI for searching logs via the main `api` service.
- Persistent data volumes for all stateful services are managed via `docker-compose` and created on the host in `/tmp/log-beacon`.
