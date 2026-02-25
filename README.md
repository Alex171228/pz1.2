# Практические занятия №1 и №2: Микросервисная архитектура

## Описание

Учебная система из двух микросервисов для демонстрации межсервисного взаимодействия через HTTP и gRPC.

### Границы сервисов

**Auth Service:**
- Выдача токенов (упрощённая модель)
- Проверка токенов
- Возвращает информацию: валидный/не валидный

**Tasks Service:**
- CRUD операции над задачами
- Перед выполнением операций проверяет токен через Auth

## Структура проекта

```
pz1.2/
├── services/
│   ├── auth/
│   │   ├── cmd/auth/main.go          # Точка входа Auth
│   │   └── internal/
│   │       ├── http/handler.go       # HTTP хендлеры
│   │       ├── grpc/server.go        # gRPC сервер
│   │       └── service/auth.go       # Бизнес-логика
│   └── tasks/
│       ├── cmd/tasks/main.go         # Точка входа Tasks
│       └── internal/
│           ├── http/handler.go       # HTTP хендлеры
│           ├── service/task.go       # Бизнес-логика
│           └── client/authclient/    # Клиенты для Auth
│               ├── client.go         # Интерфейс
│               ├── http.go           # HTTP клиент (ПЗ1)
│               └── grpc.go           # gRPC клиент (ПЗ2)
├── shared/
│   ├── middleware/
│   │   ├── requestid.go              # Middleware для X-Request-ID
│   │   └── logging.go                # Middleware для логирования
│   └── httpx/
│       └── client.go                 # HTTP клиент с таймаутом
├── proto/
│   ├── auth.proto                    # Определение gRPC контракта
│   └── auth/                         # Сгенерированный код
├── docs/
│   └── api.md                        # Документация API
├── scripts/                          # Скрипты запуска
├── go.mod
└── README.md
```

## Требования

- Go 1.21+

## Запуск на сервере

### Установка зависимостей

```bash
git clone <repository>
cd pz1.2
go mod download
```

### ПЗ1: HTTP взаимодействие

**Терминал 1 - Auth Service:**
```bash
export AUTH_PORT=8081
go run ./services/auth/cmd/auth
```

**Терминал 2 - Tasks Service (HTTP режим):**
```bash
export TASKS_PORT=8082
export AUTH_BASE_URL=http://localhost:8081
export AUTH_MODE=http
go run ./services/tasks/cmd/tasks
```

### ПЗ2: gRPC взаимодействие

**Терминал 1 - Auth Service:**
```bash
export AUTH_PORT=8081
export AUTH_GRPC_PORT=50051
go run ./services/auth/cmd/auth
```

**Терминал 2 - Tasks Service (gRPC режим):**
```bash
export TASKS_PORT=8082
export AUTH_GRPC_ADDR=localhost:50051
export AUTH_MODE=grpc
go run ./services/tasks/cmd/tasks
```

## Переменные окружения

### Auth Service

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| AUTH_PORT | Порт HTTP сервера | 8081 |
| AUTH_GRPC_PORT | Порт gRPC сервера | 50051 |

### Tasks Service

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| TASKS_PORT | Порт HTTP сервера | 8082 |
| AUTH_MODE | Режим: http или grpc | http |
| AUTH_BASE_URL | URL Auth (для HTTP) | http://localhost:8081 |
| AUTH_GRPC_ADDR | Адрес Auth (для gRPC) | localhost:50051 |

## Тестирование

### Получение токена
```bash
curl -s -X POST http://localhost:8081/v1/auth/login \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: req-001" \
  -d '{"username":"student","password":"student"}'
```

### Создание задачи
```bash
curl -i -X POST http://localhost:8082/v1/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer demo-token" \
  -H "X-Request-ID: req-002" \
  -d '{"title":"Do PZ17","description":"split services","due_date":"2026-01-10"}'
```

### Получение задач
```bash
curl -i http://localhost:8082/v1/tasks \
  -H "Authorization: Bearer demo-token" \
  -H "X-Request-ID: req-003"
```

### Запрос без токена (должен вернуть 401)
```bash
curl -i http://localhost:8082/v1/tasks \
  -H "X-Request-ID: req-004"
```

## Контрольные вопросы

### ПЗ1

1. **Почему межсервисный вызов должен иметь таймаут?**
   - Предотвращение каскадных отказов
   - Освобождение ресурсов при зависании
   - Быстрое информирование клиента об ошибке

2. **Как request-id помогает при диагностике ошибок?**
   - Позволяет связать логи разных сервисов
   - Упрощает поиск проблем в распределённой системе

3. **Какой статус нужно вернуть клиенту на невалидный токен?**
   - 401 Unauthorized

4. **Как описать "точку отказа" между сервисами?**
   - Auth недоступен → 503 Service Unavailable

### ПЗ2

1. **Что такое .proto и почему он считается контрактом?**
   - Формальное описание API
   - Генерация кода для клиента и сервера

2. **Что такое deadline в gRPC и чем он полезен?**
   - Ограничение времени на выполнение запроса
   - Автоматическая отмена при превышении

3. **Почему "exactly-once" не может быть так прост в RPC?**
   - Сетевые сбои и повторы
   - Нужна идемпотентность

4. **Как обеспечивать совместимость при расширении .proto?**
   - Добавление новых полей с новыми номерами
   - Использовать reserved для защиты
