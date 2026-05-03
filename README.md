# Inventory Hub Suite

Inventory Hub is a high-performance microservices suite designed to collect, process, and visualize hardware and software inventory data from agents (compatible with the OCS Inventory XML format).

## Architecture

The system is built as a modular pipeline:
1.  **`inv-collector` (Go):** High-speed HTTP receiver that parses incoming XML inventory payloads and produces JSON events.
2.  **`Kafka` (Apache):** Durable message broker (running in KRaft mode) that decouples ingestion from processing.
3.  **`inv-consumer` (Go):** Scalable consumer that subscribes to Kafka and persists inventory state into PostgreSQL with intelligent timestamp handling.
4.  **`PostgreSQL`:** Relational database for persistent storage.
5.  **`inv-web` (Ruby on Rails):** Web-based dashboard to browse and manage the collected inventory.

## Prerequisites

- Docker and Docker Compose
- A tool to send OCS-compatible XML data (e.g., `curl` or an OCS Agent)

## Quick Start

1.  **Configure:** Copy the example environment settings (or edit the existing `.env`):
    ```bash
    # Ensure .env exists with required keys
    RAILS_MASTER_KEY=your_key_here
    KAFKA_TOPIC=inventory_events
    KAFKA_GROUP_ID=inv-consumer-group
    ```

2.  **Build and Run:**
    ```bash
    docker compose up --build
    ```

3.  **Access:**
    - **Web Dashboard:** [http://localhost:3000](http://localhost:3000)
    - **Collector Endpoint:** [http://localhost:8088/ocsinventory](http://localhost:8088/ocsinventory)

## Configuration

The suite is primarily configured via the root `.env` file:

| Variable | Description | Default |
| :--- | :--- | :--- |
| `KAFKA_TOPIC` | The name of the Kafka topic for inventory events | `inventory_events` |
| `KAFKA_GROUP_ID` | The consumer group ID for the persistence service | `inv-consumer-group` |
| `RAILS_MASTER_KEY` | Encryption key for Rails credentials | (Required) |

Advanced field mapping can be configured in `inv-collector/config.yaml`.

## Testing

The project includes unit and integration tests for all services.

### Go Services (Collector & Consumer)
```bash
cd inv-collector && go test ./...
cd inv-consumer && go test ./...
```

### Rails Web Service
```bash
docker compose run --rm -e RAILS_ENV=test web ./bin/rails db:create db:migrate test
```

## Legacy Transition
This project replaces the older `rb-ocs-collector` and standalone scripts with a modern, containerized architecture. Legacy files (`.spec`, `rpmbuild/`, etc.) have been deprecated in favor of this Docker-native suite.
