# Log Beacon

Log Beacon is a high-performance, Humio-inspired log ingestion and search platform built with Go. It provides a simple and efficient way for developers to ingest, store, and search through application logs, aiding in debugging and analysis.

## Project Goals

- **High Performance:** Built with Go and leveraging efficient libraries like Gin for the web framework.
- **Simple & Scalable:** Designed with a clean architecture that separates log data from its metadata, allowing for cost-effective scaling.
- **Developer-Friendly:** Provides a straightforward API for log ingestion and a powerful query interface.

## Getting Started

### Prerequisites

- [Go](https://golang.org/doc/install) (version 1.18 or higher recommended)

### Running the Server

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/your-username/log-beacon.git
    cd log-beacon
    ```

2.  **Install dependencies:**
    The project uses Go Modules. Dependencies will be automatically downloaded when you run the application.

3.  **Run the server:**
    ```bash
    go run main.go
    ```
    The server will start on `http://localhost:8080`.

### API Endpoints

The following endpoints are available:

-   `GET /health`: Checks the health of the server.
-   `POST /api/v1/ingest`: The endpoint for ingesting logs.
-   `GET /api/v1/search`: The endpoint for searching logs.

#### Example Usage with cURL:

-   **Health Check:**
    ```bash
    curl http://localhost:8080/health
    ```

-   **Search:**
    ```bash
    curl "http://localhost:8080/api/v1/search?q=error"
    ```

-   **Ingest:**
    ```bash
    curl -X POST http://localhost:8080/api/v1/ingest \
         -H "Content-Type: application/json" \
         -d '{"level": "info", "message": "User logged in successfully"}'
    ```