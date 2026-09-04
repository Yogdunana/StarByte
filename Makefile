# StarByte 开发命令入口（Issue #19）
POSTGRES_USER     ?= starbyte
POSTGRES_PASSWORD ?= starbyte
POSTGRES_DB       ?= starbyte
POSTGRES_HOST     ?= localhost
POSTGRES_PORT     ?= 5432
DATABASE_URL      ?= postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable
MIGRATE_PATH      ?= backend/migrations

.PHONY: help migrate-up migrate-down migrate-create seed backend-test frontend-lint

help:
	@echo "make migrate-up              执行全部数据库迁移"
	@echo "make migrate-down            回滚最近一次迁移"
	@echo "make migrate-create name=xx  创建新的迁移文件"
	@echo "make seed                    写入幂等种子数据（可重复执行）"
	@echo "make backend-test            运行后端单测"
	@echo "make frontend-lint           前端 lint"

migrate-up:
	migrate -path $(MIGRATE_PATH) -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path $(MIGRATE_PATH) -database "$(DATABASE_URL)" down 1

migrate-create:
	@test -n "$(name)" || (echo "usage: make migrate-create name=xxx" && exit 1)
	migrate create -ext sql -dir $(MIGRATE_PATH) -seq $(name)

seed:
	cd backend && go run ./scripts

backend-test:
	cd backend && go test ./...

frontend-lint:
	cd frontend && npm run lint
