.PHONY: build run clean test test-coverage docker-build docker-run docker-stop docker-clean docker-compose-up docker-compose-down help

# Variables
APP_NAME = url-shortener
PORT = 8080
DB_PATH = data.db
BASE_URL = http://localhost:8080
DOCKER_IMAGE = url-shortener
DOCKER_CONTAINER = url-shortener

# Go build flags
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
CGO_ENABLED ?= 1

# Help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build              Build the application"
	@echo "  run                Run the application"
	@echo "  clean              Clean build artifacts"
	@echo "  test               Run tests"
	@echo "  test-coverage      Run tests with coverage"
	@echo "  docker-build       Build Docker image"
	@echo "  docker-run         Run Docker container"
	@echo "  docker-stop        Stop Docker container"
	@echo "  docker-clean       Remove Docker container and image"
	@echo "  docker-compose-up  Start with Docker Compose"
	@echo "  docker-compose-down Stop and remove Docker Compose services"
	@echo "  help               Show this help message"

# Build the application
build:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(APP_NAME) ./cmd

# Run the application
run: build
	./$(APP_NAME) --port $(PORT) --db $(DB_PATH) --base-url $(BASE_URL)

# Clean build artifacts
clean:
	rm -f $(APP_NAME)
	rm -f coverage.out

# Test commands
test:
	go test -v ./...

test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Docker commands
docker-build:
	docker build -t $(DOCKER_IMAGE) .

docker-run: docker-build
	docker run -d --name $(DOCKER_CONTAINER) \
		-p $(PORT):$(PORT) \
		-e PORT=$(PORT) \
		-e DB_PATH=/data/$(DB_PATH) \
		-e BASE_URL=$(BASE_URL) \
		-v url-shortener-data:/data \
		$(DOCKER_IMAGE)

docker-stop:
	docker stop $(DOCKER_CONTAINER) || true
	docker rm $(DOCKER_CONTAINER) || true

docker-clean: docker-stop
	docker rmi $(DOCKER_IMAGE) || true

# Docker Compose commands
docker-compose-up:
	docker-compose up -d

docker-compose-down:
	docker-compose down 