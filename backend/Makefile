PORT = 8081

.PHONY: run
run:
	@echo "Запуск приложения..."
	docker-compose up

.PHONY: swagger-ui
swagger-ui:
	@echo "Открытие Swagger UI в браузере..."
	open http://localhost:$(PORT)/swagger/index.html

.PHONY: help
help:
	@echo "Доступные команды:"
	@echo "  make run          - Запустить приложение"
	@echo "  make swagger-ui   - Открыть Swagger UI в браузере"