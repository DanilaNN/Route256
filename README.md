- Проект состоит из трёх сервисов, имитирующих обработку заказа в интернет-магазине
- API взаимодействия сервисов в файле contracts.md
- Общение сервисов по http-json-rpc
- Информация о заказах хранится в PostgreSQL
- Изменения статусов заказов сохраняются в Kafka и отсылаются в Telegram
- make run-all запускает Docker контейнеры с сервисами и инфраструктурой для метрик, трейсинга и логирования.

- prometheus/grafana - обработка, отображение метрик
- jaeger - визуализация трассировок
- TODO: graylog - сбор, обработка логов
