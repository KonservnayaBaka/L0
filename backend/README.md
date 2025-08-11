# Запуск и документация проекта

##  Сервисы и адреса

- **Swagger UI (API документация)** — [http://localhost:8081/swagger/index.html](http://localhost:8081/swagger/index.html)
- **Prometheus (метрики)** — [http://localhost:9090/](http://localhost:9090/)
- **Grafana (дашборды)** — [http://localhost:3000/](http://localhost:3000/)  
  **Логин/Пароль**: `admin` / `admin`

---

## Запуск приложения

Запустить проект можно только через **`docker-compose`** :

   ```bash
   docker-compose up
   ```

Сервис доступен по порту 8081

---

## Используемые технологии
- **Gin** - Веб фреимворк
- **PostgreSQL** - Основная база данных проекта
- **Redis** - Кеш
- **Promiteus** - Сбор метрик
- **Grafana** - Сбор логов
- **Swagger** - API документация
