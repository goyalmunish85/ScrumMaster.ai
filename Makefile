.PHONY: up down dev-db init-backend init-frontend

up:
	docker-compose up -d

down:
	docker-compose down

dev-db:
	docker-compose up -d postgres redis qdrant

init-backend:
	mkdir -p backend/cmd/api backend/cmd/worker backend/internal/common backend/internal/modules/identity backend/internal/modules/chat backend/internal/modules/ops backend/internal/modules/events backend/internal/modules/memory backend/internal/modules/ai backend/pkg backend/migrations
	cd backend && go mod init github.com/aios/backend

