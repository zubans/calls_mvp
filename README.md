# Сервер для видеозвонков на Go

Этот проект реализует сервер для видеозвонков с использованием WebRTC и языка программирования Go.

## Возможности

- Создание комнат для видеозвонков
- Подключение нескольких участников к одной комнате
- Обмен сигнальными сообщениями через WebSocket
- Поддержка ICE кандидатов и SDP описаний
- Graceful shutdown сервера
- **WebSocket поддержка** для более эффективного обмена сигнальными сообщениями
- **Аутентификация пользователей** с использованием JWT токенов
- **Запись звонков** в формате WebM
- **Чат в комнате** для текстового общения
- **Мониторинг и метрики** с помощью Prometheus

## Требования

- Go 1.19 или выше
- Интернет-соединение для STUN сервера

## Установка

1. Клонируйте репозиторий:
   ```bash
   git clone https://github.com/zubans/video-call-server.git
   cd video-call-server
   ```

2. Установите зависимости:
   ```bash
   go mod tidy
   ```

## Запуск

1. Запустите сервер:
   ```bash
   go run main.go
   ```

2. Сервер будет доступен по адресу `http://localhost:8080`

## Использование

### Регистрация пользователя

```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "user1",
    "email": "user1@example.com",
    "password": "password123"
  }'
```

### Вход в систему

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "user1",
    "password": "password123"
  }'
```

### Создание комнаты

```bash
curl -X POST http://localhost:8080/create-room \
  -H "Content-Type: application/json" \
  -H "Authorization: YOUR_JWT_TOKEN" \
  -d '{"name": "my-room"}'
```

### Присоединение к комнате

```bash
curl -X POST http://localhost:8080/join-room \
  -H "Content-Type: application/json" \
  -H "Authorization: YOUR_JWT_TOKEN" \
  -d '{"room_id": "room_1234567890"}'
```

### Отправка сообщения в чат

```bash
curl -X POST http://localhost:8080/chat/send \
  -H "Content-Type: application/json" \
  -H "Authorization: YOUR_JWT_TOKEN" \
  -d '{
    "room_id": "room_1234567890",
    "message": "Hello, everyone!"
  }'
```

### Начало записи звонка

```bash
curl -X POST http://localhost:8080/recording/start \
  -H "Content-Type: application/json" \
  -H "Authorization: YOUR_JWT_TOKEN" \
  -d '{"room_id": "room_1234567890"}'
```

### Остановка записи звонка

```bash
curl -X POST http://localhost:8080/recording/stop \
  -H "Content-Type: application/json" \
  -H "Authorization: YOUR_JWT_TOKEN" \
  -d '{"recording_id": "recording_1234567890"}'
```

## Тестирование с помощью Postman

В директории `postman` вы найдете коллекцию Postman для тестирования API сервера. Импортируйте файл `VideoCallServer.postman_collection.json` в Postman для удобного тестирования всех endpoint'ов.

## API Endpoints

- `POST /register` - Регистрация нового пользователя
- `POST /login` - Вход в систему
- `GET /health` - Проверка состояния сервера

Защищенные endpoints (требуют JWT токен в заголовке Authorization):
- `POST /create-room` - Создание новой комнаты
- `POST /join-room` - Присоединение клиента к комнате
- `POST /leave-room` - Отключение клиента от комнаты
- `GET /rooms` - Получение списка активных комнат
- `GET /ws` - WebSocket соединение для сигнальных сообщений
- `POST /chat/send` - Отправка сообщения в чат
- `GET /chat/history/:room_id` - Получение истории чата комнаты
- `POST /recording/start` - Начало записи звонка
- `POST /recording/stop` - Остановка записи звонка
- `GET /recording/list/:room_id` - Получение списка записей комнаты
- `GET /metrics` - Метрики Prometheus

## Архитектура

Сервер состоит из следующих компонентов:

1. **RoomManager** - управляет всеми комнатами для видеозвонков
2. **UserManager** - управляет пользователями и аутентификацией
3. **ChatManager** - управляет сообщениями чата
4. **RecordingManager** - управляет записями звонков
5. **WebSocket Hub** - управляет WebSocket соединениями
6. **Metrics** - собирает и предоставляет метрики для мониторинга

## Лицензия

MIT