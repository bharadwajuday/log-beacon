# Kubernetes Deployment Guide for Log Beacon

This directory contains production-ready Kubernetes specifications to deploy Log Beacon to any Kubernetes cluster (Minikube, Kind, k3s, EKS, GKE, AKS, or bare metal).

---

## Architecture & Workload Types

| Service | Type | Replicas | Rationale |
| :--- | :--- | :---: | :--- |
| **`postgres`** | `StatefulSet` | **1** | RDBMS with data on disk; requires volume persistence and single-instance writer. |
| **`nats`** | `StatefulSet` | **1** | Persistent NATS JetStream storage on dedicated disk volume. |
| **`minio`** | `StatefulSet` | **1** | S3-compatible cold object storage holding Apache Parquet archives. |
| **`hot-storage`** | `StatefulSet` | **1** | In-memory/embedded Bleve and BadgerDB with exclusive directory locking. |
| **`archiver`** | `Deployment` | **1** (can scale) | Queue worker buffering incoming NATS logs and writing Parquet batches to MinIO. |
| **`api`** | `Deployment` | **2+** | Stateless HTTP and WebSocket API Gateway. |
| **`frontendv2`** | `Deployment` | **2+** | Nginx serving React SPA and reverse-proxying `/api/` to `api:8080`. |

---

## File Overview

* [`00-namespace.yaml`](00-namespace.yaml): Creates the isolated `log-beacon` namespace.
* [`01-configmaps-secrets.yaml`](01-configmaps-secrets.yaml): Environment configuration and credentials (DB URL, MinIO credentials, batch parameters).
* [`10-postgres.yaml`](10-postgres.yaml): Postgres StatefulSet, headless service, PVC, and `init.sql` ConfigMap.
* [`20-nats.yaml`](20-nats.yaml): NATS JetStream StatefulSet with persistent JetStream buffer.
* [`30-minio.yaml`](30-minio.yaml): MinIO object storage StatefulSet, Service, and PVC.
* [`40-hot-storage.yaml`](40-hot-storage.yaml): Hot Storage StatefulSet, Service, and PVC with background retention pruner.
* [`50-archiver.yaml`](50-archiver.yaml): Archiver queue worker with Parquet batching.
* [`60-api.yaml`](60-api.yaml): API Gateway Deployment and Service with federated search routing.
* [`70-frontend.yaml`](70-frontend.yaml): React frontend Deployment and Service.
* [`kustomization.yaml`](kustomization.yaml): Kustomize file bundling all resources.

---

## Quickstart Deployment

### 1. Build & Push Docker Images
Ensure container images are tagged and accessible to your Kubernetes cluster:
```bash
# Build backend images
docker build -t log-beacon:latest .

# Build frontend image
docker build -t log-beacon-frontendv2:latest ./frontendv2

# (If using Minikube / Kind)
minikube image load log-beacon:latest
minikube image load log-beacon-frontendv2:latest
# or
kind load docker-image log-beacon:latest
kind load docker-image log-beacon-frontendv2:latest
```

### 2. Deploy Manifests
Deploy all services in a single command using Kustomize:
```bash
kubectl apply -k k8s/
```
Or with plain `kubectl`:
```bash
kubectl apply -f k8s/
```

### 3. Verify Pod Status
```bash
kubectl get pods -n log-beacon
```
Expected output:
```
NAME                          READY   STATUS    RESTARTS   AGE
api-648b7654bf-abcde          1/1     Running   0          30s
api-648b7654bf-fghij          1/1     Running   0          30s
archiver-594dfc7f76-klmno     1/1     Running   0          30s
frontendv2-74898fc66b-pqrst   1/1     Running   0          30s
frontendv2-74898fc66b-uvwxy   1/1     Running   0          30s
hot-storage-0                 1/1     Running   0          30s
minio-0                       1/1     Running   0          30s
nats-0                        1/1     Running   0          30s
postgres-0                    1/1     Running   0          30s
```

### 4. Access the Application
Forward the frontend service to your local machine:
```bash
kubectl port-forward svc/frontendv2 3000:80 -n log-beacon
```
Open `http://localhost:3000` in your browser.

To access the MinIO console:
```bash
kubectl port-forward svc/minio 9001:9001 -n log-beacon
```
Open `http://localhost:9001` (User: `minioadmin`, Pass: `minioadmin`).
