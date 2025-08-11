PORT = 8081

.PHONY: run
run:
	@echo "Запуск приложения..."
	docker-compose up

.PHONY: swagger-ui
swagger-ui:
	@echo "Открытие Swagger UI..."
	open http://localhost:$(PORT)/swagger/index.html

prometheus:
	@echo "Открытие Prometheus..."
	open http://localhost:9090

grafana:
	@echo "Открытие Grafana..."
	open http://localhost:3000

test:
	@echo "Запуск go test"
	@go test ./... -v

.PHONY: help
help:
	@echo "Доступные команды:"
	@echo "  make run          - Запустить приложение"
	@echo "  make swagger-ui   - Открыть Swagger UI в браузере"