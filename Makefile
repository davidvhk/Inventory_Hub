# Inventory Hub Debugging Makefile

include .env
export $(shell sed 's/=.*//' .env)

KAFKA_CONTAINER=davidvhk-inventory-overview-kafka-1
POSTGRES_CONTAINER=davidvhk-inventory-overview-postgres-1
TOPIC=$(KAFKA_TOPIC)

.PHONY: help kafka-read db-shell db-query logs collector-logs consumer-logs web-logs

help:
	@echo "Inventory Hub Debugging Commands:"
	@echo "  make kafka-read      - Read messages from Kafka topic ($(TOPIC))"
	@echo "  make db-shell        - Open PostgreSQL interactive shell"
	@echo "  make db-query        - List all devices in the database"
	@echo "  make logs            - Tail logs from all services"
	@echo "  make collector-logs  - Tail logs from the collector"
	@echo "  make consumer-logs   - Tail logs from the consumer"
	@echo "  make web-logs        - Tail logs from the web dashboard"

kafka-read:
	docker exec -it $(KAFKA_CONTAINER) /opt/kafka/bin/kafka-console-consumer.sh \
		--bootstrap-server localhost:9092 \
		--topic $(TOPIC) \
		--from-beginning \
		--max-messages 10

db-shell:
	docker exec -it $(POSTGRES_CONTAINER) psql -U inv_user -d inventory_db

db-query:
	docker exec -it $(POSTGRES_CONTAINER) psql -U inv_user -d inventory_db -c "SELECT * FROM devices;"

logs:
	docker compose logs -f

collector-logs:
	docker compose logs -f collector

consumer-logs:
	docker compose logs -f consumer

web-logs:
	docker compose logs -f web
