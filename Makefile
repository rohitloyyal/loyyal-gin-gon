TOP_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
COMPOSE := docker-compose -f $(TOP_DIR)/config/docker-compose.yaml  --project-directory $(TOP_DIR)


start:
	@cd config && $(COMPOSE) up -d api cashier retry
	@cd config && $(COMPOSE) up -d couch jaeger1 nats 
	@echo "Environment up

stop:
	@cd config && $(COMPOSE) down --remove-orphans
	@echo "Environment down

clean:
	@docker volume prune -f  2>/dev/null || echo "no volumes to prune"

purge:
	@docker rmi -f `docker images | grep "827830277284.dkr.ecr.me-south-1.amazonaws.com" | cut -d ' ' -f 1 2>/dev/null` || true

build:
	buildapi cashier retry

buildapi:
	@echo "Building api ..."
	@./scripts/dockerbuild.sh ./cmd/api
	@echo "Building callback ..."
	@./scripts/dockerbuild.sh ./cmd/callback

cashier: certs
	@echo "Building cashier ..."
	@./scripts/dockerbuild.sh ./cmd/cashier .

retry:
	@echo "Building retry ..."
	@./scripts/dockerbuild.sh ./cmd/retry