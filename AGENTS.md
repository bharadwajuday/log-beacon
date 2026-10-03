# Multi-Agent Coordination Rules & Guidelines

This document outlines the rules, guidelines, and coordination practices that multiple AI agents (or developers) must follow when modifying the Log Beacon codebase concurrently. Following these rules prevents merge conflicts, regressions, and interface mismatches.

## 1. Directory & Service Ownership
Log Beacon is split into separate microservices and a frontend application. To prevent concurrency conflicts:
* **Scope boundaries:** Limit your file edits to the service directory you are specifically tasked with:
  * **API Gateway / Ingestion:** `main.go`, `internal/server/`, `internal/auth/`, `internal/repository/`.
  * **Hot Storage:** `cmd/hot-storage/`.
  * **Archiver:** `cmd/archiver/`.
  * **Frontend:** `frontendv2/`.
* **Private internals:** Under `cmd/`, services use their own `internal/` packages (e.g., `cmd/hot-storage/internal/search`). Keep logic specific to one service in these sub-internal folders to avoid cluttering or breaking shared root-level `internal/` packages.

## 2. API Contract & Schema Coordination
If you are modifying endpoints or communication schemas between services:
1. **Do not modify interfaces without coordination:** If you need to change a shared schema (e.g., `internal/model/log.go`), verify if other agents are working on components that rely on the old structure.
2. **Backward Compatibility:** When adding API parameters, make them optional by default so that older components do not break before they are updated.
3. **Frontend-Backend Sync:** When updating API routes or request/response formats in the backend `api` server, document the changes clearly or coordinate with the agent working on `frontendv2` to update the React client synchronously.

## 3. Database & Shared State Modifications
* **Postgres Database Schema:** Initial SQL is managed via [init.sql](file:///Users/bharadwajuday/Study/projects/log-beacon/db/init.sql). Any changes to tables or authentication schema must be done carefully to avoid breaking user registration/login for others.
* **NATS Events (`log.events`):** The message publisher and subscribers must agree on the serialization format. Avoid breaking changes to the serialization of `model.Log` on NATS.

## 4. Concurrent Testing and Port Management
* **Test Isolation:** The test suite is invoked via `make test`, which starts an isolated test Docker compose setup (`docker-compose.test.yml`). 
* **Port Conflicts:** Since tests bind to local host ports (e.g. Postgres on 5432, NATS on 4222), multiple agents running `make test` concurrently on the same machine will experience port conflicts. Ensure you check for running test environments or wait for other agents' tests to complete before starting yours.

## 5. Coding Standards & Code Integrity
* **Unit Testability:** When adding new packages or modifying existing Go logic, ensure all code is covered by unit tests. Run `make test` to verify your changes.
* **Preserve Code Style & Comments:** Do not remove unrelated comments, documentation, or structure. Keep functions small, testable, and documented.
* **Update Documentation:** If you introduce new features, configuration environment variables, or schema updates, remember to update `README.md` and `AGENT.md` accordingly.
