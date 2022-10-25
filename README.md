# Go-nats
Реализация сервиса для получения данных от NATS Streaming, и вывод полученных данных в интерфейс.

В сервисе реализовано следующее
- Подключение и подписка на канал в nats-streaming
- Полученные данные сохраняются в Postgres
- Так же полученные данные сохранаются in memory в сервисе (Кеш)
- В случае падения сервиса Кеш восстанавливается из Postgres
- Данные выдаются по id из кеша
- Реализован интерфейс отображения полученных данных, для их запроса по id
