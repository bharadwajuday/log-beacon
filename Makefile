# Makefile for the Log Beacon project

# Use .PHONY to ensure these targets run even if files with the same name exist.
.PHONY: up down logs clean

# Define variables
COMPOSE_FILE := docker-compose.yml
HOST_DATA_DIRS := /tmp/log-beacon/nats-data /tmp/log-beacon/minio-data /tmp/log-beacon/hot-storage-data

# Default target: running 'make' will be the same as 'make up'
default: up

# Target to build and start all services in detached mode.
up:
	@echo "--> Ensuring host directories for persistent data exist..."
	@mkdir -p $(HOST_DATA_DIRS)
	@echo "--> Building and starting services in detached mode..."
	@docker-compose -f $(COMPOSE_FILE) up --build -d

# Target to stop and remove all services.
down:
	@echo "--> Stopping and removing all services..."
	@docker-compose -f $(COMPOSE_FILE) down

# Target to follow the logs of all running services.
logs:
	@echo "--> Tailing logs for all services (Press Ctrl+C to stop)..."
	@docker-compose -f $(COMPOSE_FILE) logs -f

# Target to clean up the host data directories.
clean:
	@echo "--> Removing host data directories..."
	@rm -rf /tmp/log-beacon
