b-build:
	@echo "Building backend development version..."
	docker build -t serra-server:latest-dev -f ./backend/Dockerfile ./backend

f-build:
	@echo "Building frontend development  version..."
	docker build -t serra-frontend:latest-dev -f ./frontend/Dockerfile ./frontend

up:
	@echo "Starting docker compose up..."
	docker-compose -f docker-compose.yml up --build -d

down:
	@echo "Starting docker compose down..."
	docker-compose -f docker-compose.yml down