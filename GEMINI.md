# Gemini Assistant Context

This document provides context for AI assistants working on the Log Beacon project.

## Project Overview

Log Beacon is a log management system inspired by Humio. The primary goal is to build a performant and scalable backend service for ingesting, storing, and searching logs.

## Tech Stack

-   **Language:** Go
-   **Web Framework:** Gin (`github.com/gin-gonic/gin`)
-   **Backend Focus:** The current development focus is purely on the backend. No frontend work is planned at this stage.

## High-Level Architecture

-   **Ingestion API:** A RESTful endpoint (`/api/v1/ingest`) to receive log data.
-   **Search API:** A RESTful endpoint (`/api/v1/search`) to query logs.
-   **Storage:**
    -   **Index (Metadata):** An embedded key-value store like BadgerDB or BoltDB is planned.
    -   **Log Data (Chunks):** A local filesystem or an object store (like MinIO) is planned.
-   **Design Philosophy:** The system separates log metadata (labels/tags) from the raw log message. The index is used for fast filtering, and then a full-text search is performed on a smaller subset of log data chunks.

## Current Status

-   A basic Gin server has been set up in `main.go`.
-   Placeholder handlers for the ingest and search endpoints have been created.
-   The project has been initialized as a Go module.
