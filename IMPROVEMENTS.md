# Log Beacon: Architecture & Innovation Improvements

This document captures high-value, cool, and innovative engineering improvements to elevate Log Beacon from a standard log collector into a modern, production-grade observability platform.

---

## 1. Breakthrough Architecture & Storage Innovations

### 1.1 Columnar Cold Tier with DuckDB & Parquet
* **Current State:** The Archiver writes compressed raw JSON log blocks directly to MinIO. Querying these files requires slow, manual decompression and sequential scanning.
* **Proposed Enhancement:** 
  * Convert batched logs into **Apache Parquet** format partitioned by time and service (e.g., `s3://logs/year=2026/month=09/day=29/service=auth/*.parquet`).
  * Embed an in-process analytical engine (**DuckDB**) into the backend to query cold S3/MinIO files directly using SQL (`SELECT * FROM read_parquet('s3://logs/**/*.parquet') WHERE level = 'ERROR'`).
  * **Impact:** 10x–50x faster analytical queries over historical logs without incurring Elasticsearch memory overhead.

### 1.2 Time-Windowed Hot Storage Shards & Pruning
* **Current State:** Hot storage writes continuously into a single Bleve index and BadgerDB instance, risking unbounded disk growth and query degradation over time.
* **Proposed Enhancement:**
  * Partition hot indexes into sliding time windows (e.g., 6-hour or 24-hour shards).
  * Automatically drop or unmount expired shards according to a configurable retention policy (e.g., retain 48h hot logs).
  * Hot-to-cold handoff: verify that a chunk is archived in MinIO before dropping the local Bleve shard.

---

## 2. Next-Gen Search & Analytics Capabilities

### 2.1 Aggregations & Log Volume Histograms
* **Overview:** Provide quick visual histograms showing log distribution and error spikes over time.
* **Key Features:**
  * Add a bucketed count endpoint (`GET /api/v1/search/histogram?interval=1m&query=...`).
  * Display a clickable bar chart above search results in the frontend allowing users to zoom into time spikes.

### 2.2 Rich Query Syntax & Field Extraction
* **Overview:** Expand search beyond simple equality and `AND` operators.
* **Features:**
  * Support operators: `OR`, `NOT`, wildcards (`service:user-*`), numeric comparisons (`duration_ms > 500`), and regex.
  * Dynamic JSON parsing: Allow searching nested JSON payload fields without upfront schema declaration (e.g., `json.http.status:500`).

---

## 3. AI-Assisted Observability & Root Cause Analysis

### 3.1 Automatic Log Clustering & Anomaly Detection
* **Problem:** During an outage, thousands of duplicate error logs flood the system, burying the root cause.
* **Proposed Enhancement:**
  * Implement Drain or semantic clustering to group repeating log patterns into message templates (e.g., `Failed to connect to db host <*>, retry in <*>` count: 4,821).
  * Highlight "novel" log lines that have never been observed before in the given service.

### 3.2 One-Click Incident Summarization
* **UI Action:** A "Summarize Errors" button in the frontend search view.
* **Mechanism:** Batches matching error logs within the active time window and invokes an LLM to generate an incident post-mortem summary, pinpointing root cause services and affected endpoints.

---

## 4. Developer Experience & Production Readiness

### 4.1 Ingestion Protocols & Agent Integrations
* Support standard ingestion protocols to make Log Beacon a drop-in replacement:
  * OpenTelemetry (OTel) Collector exporter / OTLP HTTP receiver (`/v1/logs`).
  * Vector / FluentBit / Logstash HTTP sink compatibility.
  * API Key / Service Account token authentication for `/api/v1/ingest`.

### 4.2 UI / UX Modernization
* **Search Polish:** Search history autocomplete, saved queries/favorites, debounced input, and query syntax validation.
* **Log Detail Drawer:** Slide-out drawer displaying JSON formatting, copy-as-cURL, and label filtering shortcuts.
* **Live Tail Controls:** Pause/Resume stream button, buffer limit controls (e.g. keep last 1000 lines), and regex search inside live streams.
