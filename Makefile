include .env
export $(shell sed 's/=.*//' .env)


#============Docker============
SERVICE=app
# Собрать образы
build:
	docker-compose build $(SERVICE)

# Запустить контейнеры (в фоне) с пересборкой образа
up:
	docker-compose up -d --build --force-recreate $(SERVICE)

# Остановить контейнеры
down:
	docker-compose down

# Перезапустить контейнер (с пересборкой и пересозданием)
restart: down up

# Просмотр логов сервиса
logs:
	docker-compose logs -f $(SERVICE)


#============МИГРАЦИИ============

goose-install:
	go install github.com/pressly/goose/v3/cmd/goose@latest

goose-add:
	goose -dir ./migrations postgres "$(DATABASE_DSN_MIGRATIONS)" create rename_me sql

goose-up:
	goose -dir ./migrations postgres "$(DATABASE_DSN_MIGRATIONS)" up

goose-down:
	goose -dir ./migrations postgres "$(DATABASE_DSN_MIGRATIONS)" down

goose-status:
	goose -dir ./migrations postgres "$(DATABASE_DSN_MIGRATIONS)" status


#============Моки============
mock-generate:
	mockgen -source="internal/advertisement/service.go" -destination="internal/advertisement/mock/mock_repository_interface.go" -package=mockad
	mockgen -source="internal/advertisement/handler.go" -destination="internal/advertisement/mock/mock_service_interface.go" -package=mockad

	mockgen -source="internal/user/service.go" -destination="internal/user/mock/mock_repository_interface.go" -package=mockuser
	mockgen -source="internal/user/handler.go" -destination="internal/user/mock/mock_service_interface.go" -package=mockuser

#============Тесты============
test:
	go test -cover ./internal/advertisement
	go test -cover ./internal/user

test-ad:
	go test -cover ./internal/advertisement -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

test-user:
	go test -cover ./internal/user -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

#============Swagger============
swag-gen:
	swag init -g cmd/main.go
