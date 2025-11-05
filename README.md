# Log Beacon

Log Beacon is a high-performance, Humio-inspired log ingestion and search platform built with Go. It provides a simple and efficient way for developers to ingest, store, and search through application logs, aiding in debugging and analysis.

## Project Goals

- **High Performance:** Built with Go and leveraging efficient libraries like Gin for the web framework.
- **Simple & Scalable:** Designed with a clean architecture that separates log data from its metadata, allowing for cost-effective scaling.
- **Developer-Friendly:** Provides a straightforward API for log ingestion and a powerful query interface.

## Project Structure

The project is organized into several key packages:

- `cmd/`: Contains the `main` packages for the different binaries (`api` and `consumer`).
- `internal/`: Contains the core business logic for the application.
  - `server/`: The Gin web server and API handlers.
  - `consumer/`: The NATS message consumer and log processor.
  - `queue/`: NATS connection and stream management logic.
  - `model/`: Core data structures like the `Log` entry.
- `docker-compose.yml`: Defines the services, networks, and volumes for the local development environment.
- `Dockerfile`: A multi-stage Dockerfile for building lean, production-ready images of the Go applications.

## Getting Started

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/)
- [Make](https://www.gnu.org/software/make/)

### Running the Environment

The entire development environment is managed via a `Makefile` for simplicity and consistency.

1. **Clone the repository:**

    ```bash
    git clone https://github.com/your-username/log-beacon.git
    cd log-beacon
    ```

2. **Build and Run Services:**
    This command ensures the necessary host directories are created, then builds the images and starts all services (`nats`, `minio`, `api`, `archiver`, `hot-storage`) in detached mode.

    ```bash
    make
    ```

    The API server will be available at `http://localhost:8080`, the MinIO console at `http://localhost:9001`, and the web UI at `http://localhost:3000`.

3. **Follow Logs:**
    To view the real-time logs from all running services:

    ```bash
    make logs
    ```

4. **Stop Services:**
    To stop and remove all running containers:

    ```bash
    make down
    ```

5. **Clean Up Data:**
    To remove the persistent data stored on the host machine in `/tmp/log-beacon`:

    ```bash
    make clean
    ```

### API Endpoints

The following endpoints are available:

- `GET /health`: Checks the health of the server.
- `POST /api/v1/ingest`: The endpoint for ingesting logs.
- `GET /api/v1/search`: The endpoint for searching logs.

#### Example Usage with cURL

- **Health Check:**

    ```bash
    curl http://localhost:8080/health
    ```

- **Search:**

    ```bash
    curl "http://localhost:8080/api/v1/search?q=error"
    ```

- **Ingest:**

    ```bash
    curl -X POST http://localhost:8080/api/v1/ingest \
         -H "Content-Type: application/json" \
         -d '{"level": "info", "message": "User logged in successfully"}'
    ```
