# Log Beacon

Log Beacon is a high-performance, Humio-inspired log ingestion and search platform built with Go and React. It provides a simple and efficient way for developers to ingest, store, and search through application logs, aiding in debugging and analysis.

## Features

- **Log Ingestion**: HTTP API for ingesting logs.
- **Hot Storage**: Fast, indexed search using Bleve and BadgerDB.
- **Cold Storage**: Long-term archival to MinIO.
- **Search**:
    - Full-text search on log messages.
    - Structured search on fields (e.g., `level:error`, `service:auth`).
    - Boolean operators: `AND`, `OR` (e.g., `service:auth AND level:error`).
    - Log level filtering via UI.
- **Frontend**: Modern React-based UI for searching and viewing logs.

## Architecture

Log Beacon uses a decoupled, microservices-oriented architecture built on a hot/cold storage strategy.

- **Frontend (`frontendv2`):** A React single-page application providing the search UI, built with Vite and Tailwind CSS.
- **API Handler (`api`):** The main entrypoint for ingestion and search requests.
- **Message Queue (`nats`):** A NATS server with JetStream that acts as a durable buffer for incoming logs.
- **Hot Storage (`hot-storage`):** A consumer that indexes recent logs in Bleve and BadgerDB for fast, real-time searching.
- **Cold Storage (`archiver`):** A consumer that archives all logs to a MinIO object store for long-term retention.
- **Object Storage (`minio`):** A MinIO server for durable, long-term log archival.

## Getting Started

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/)
- [Make](https://www.gnu.org/software/make/)

### Running the Environment

The entire development environment is managed via a `Makefile` for simplicity and consistency.

1. **Clone the repository:**

    ```bash
    git clone https://github.com/bharadwajuday/log-beacon.git
    cd log-beacon
    ```

2. **Build and Run Services:**
    This command builds the images and starts all services in detached mode.

    ```bash
    make
    ```

3. **Access Services:**
    - **Web UI:** `http://localhost:3000`
    - **API Server:** `http://localhost:8080`
    - **MinIO Console:** `http://localhost:9001` (user: `minioadmin`, pass: `minioadmin`)

### Usage

- **Ingest Logs:** Send logs to the `/api/v1/ingest` endpoint.

    ```bash
    curl -X POST http://localhost:8080/api/v1/ingest -H "Content-Type: application/json" -d '{"message": "User authentication failed"}'
    ```

- **Search Logs:** Use the web UI at `http://localhost:3000`.

### Managing the Environment

- **Follow Logs:**

    ```bash
    make logs
    ```

- **Stop Services:**

    ```bash
    make down
    ```

- **Clean Up Data:**
    To remove the persistent data stored on the host machine in `/tmp/log-beacon`:

    ```bash
    make clean
    ```

- **Run Tests:**

    ```bash
    make test
    ```
