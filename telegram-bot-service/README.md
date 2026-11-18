# telegram-bot-service

Независимый микросервис, который предоставляет доступ к функциям платформы Finance Microservices через Telegram. Сервис не хранит финансовые операции, бюджеты и списки покупок и не реализует бизнес-логику повторно: все операции выполняются через существующий `gateway-service` и доменные gRPC-сервисы платформы.

## Назначение

`telegram-bot-service` решает четыре задачи:

1. Принимает Telegram updates через Webhook или Long Polling.
2. Связывает Telegram-пользователя с существующим аккаунтом Finance Microservices через одноразовый код.
3. Выступает Telegram-интерфейсом к финансам, спискам покупок, AI-планировщику, аналитике и валютам.
4. Доставляет пользователю уведомления из Kafka и предоставляет Telegram Web App для интерактивной карты AI-плана.

Сервис использует собственные PostgreSQL и Redis. PostgreSQL хранит только Telegram-привязки, одноразовые коды и настройки уведомлений. Redis хранит состояния диалогов и ключи идемпотентности.

## Архитектура

```mermaid
flowchart TB
    Telegram[Telegram Bot API]
    Webhook[Webhook / Long Polling]
    Bot[telegram-bot-service]
    Redis[(Redis)]
    PG[(Telegram PostgreSQL)]
    Kafka[(Kafka finance-events)]
    Gateway[gateway-service]
    Auth[auth-service]
    Users[users-service]
    Finance[finance-service]
    Shopping[shopping-list-service]
    AI[ai-service]
    Currency[currency-service]

    Telegram -->|HTTPS updates| Webhook
    Webhook --> Bot
    Bot --> PG
    Bot --> Redis
    Kafka -->|budget.exceeded, ai.plan_created, transaction.created, user.deleted| Bot
    Bot -->|REST + JWT| Gateway
    Gateway -->|gRPC| Auth
    Gateway -->|gRPC| Users
    Gateway -->|gRPC| Finance
    Gateway -->|gRPC| Shopping
    Gateway -->|gRPC| AI
    Gateway -->|gRPC| Currency
    Bot -->|Telegram API| Telegram
```

### Принципы интеграции

- Telegram-бот не обращается к PostgreSQL других сервисов.
- Все пользовательские операции выполняются через `gateway-service` с JWT пользователя.
- JWT выпускается сервисом на короткий срок только после разрешения привязки Telegram ID → user ID.
- Идентификатор пользователя никогда не берётся из текста Telegram-сообщения.
- AI-запросы проходят только через `ai-service`; бот не обращается к Groq, Nominatim, OSRM или другим провайдерам.
- Kafka используется только для получения событий и доставки уведомлений.

### Слои

```text
Telegram Update
      |
      v
transport/telegram  (маршрутизация, клавиатуры, диалоги, форматирование)
      |
      v
features/service     (привязка, состояния, уведомления, идемпотентность)
      |
      v
features/repository  (PostgreSQL и Redis адаптеры)
      |
      v
core/clients/gateway (REST-клиент существующего API)
```

## Структура каталогов

```text
telegram-bot-service/
├── cmd/telegram/main.go
├── config/config.go
├── internal/core/
│   ├── auth/jwt/
│   ├── cache/
│   ├── clients/gateway/
│   ├── domain/
│   ├── grpc/
│   ├── kafka/
│   ├── metrics/
│   ├── ports/
│   ├── repository/postgres/pool/{pgx}/
│   └── telemetry/
├── internal/features/
│   ├── repository/postgres/
│   ├── repository/redis/
│   ├── service/
│   ├── transport/telegram/
│   ├── transport/grpc/
│   ├── transport/http/
│   ├── transport/kafka/
│   └── webapp/
├── migrations/
├── pkg/{errors,grpcutil,logger,pagination}/
├── proto/telegram/v1/telegram.proto
├── proto/telegram/gen/
├── webapp assets are embedded from internal/features/webapp/assets
├── Dockerfile
├── go.mod
├── go.sum
└── .env.example
```

## Функции

| Раздел | Возможности | Статус |
|---|---|---|
| Привязка аккаунта | `/start КОД`, одноразовый код, TTL, запрет повторной привязки, `/unlink`, отзыв через web API | Реализовано и проверено тестами |
| Главное меню | Inline Keyboard, возврат, защита от повторных нажатий | Реализовано и проверено тестами |
| Финансы | Баланс, доходы, расходы, последние операции, категории, фильтры по типу/категории/датам, CRUD операций, описание операции | Реализовано; сценарий создания операции проверен тестами |
| Бюджеты | Список, прогресс, остаток, создание, изменение, удаление | Реализовано |
| Аналитика | Периоды: день, неделя, месяц, год; доходы, расходы, итог, топ категорий | Реализовано |
| Экспорт | CSV, PDF, JSON, TXT; отправка файла в Telegram | Реализовано |
| Списки покупок | Список, пагинация, создание, просмотр, изменение, удаление, завершение, прогресс, просмотр и загрузка изображения | Реализовано |
| AI-планировщик | Текст запроса, адрес, геолокация Telegram, координаты, магазины, цены, наличие, доставка, маршрут, рекомендация, история, удаление | Реализовано |
| Telegram Web App | Интерактивная карта AI-плана, проверка `initData`, короткоживущий JWT | Реализовано; проверка `initData` и сессия проверены тестами |
| Валюты | Курсы, базовые валюты, конвертация | Реализовано |
| Уведомления | `budget.exceeded`, `ai.plan_created`, `transaction.created`, `user.deleted`; настройки; повторные попытки | Реализовано и проверено тестами |
| Webhook | Секретный заголовок, защита от повторной обработки, очередь | Реализовано и проверено тестами |
| Long Polling | Режим локальной разработки | Реализовано |

### Ограничения существующего API

- В `shopping-list-service` список покупок — это финансовый список с полями `title`, `description`, `amount_now`, `amount_finish`, а не набор отдельных товаров. Поэтому «добавление товара» и «отметка товара» реализованы как изменение сумм и отметка всего списка выполненным. Существующий API не поддерживает отдельные позиции.
- Обновление списка в `shopping-list-service` требует положительного значения `amount_now`; при нулевом накоплении бот сообщает об ограничении и предлагает указать сумму.
- Уведомления по `transaction.updated` и `transaction.deleted` не отправляются, чтобы не создавать шум. Обрабатывается `transaction.created`.

## Технологии

- Go 1.26.4
- `github.com/go-telegram/bot` для Telegram Bot API
- PostgreSQL 15 через `pgx/v5`
- Redis 7 через `redis/go-redis/v9`
- Kafka через `segmentio/kafka-go`
- gRPC и Protocol Buffers для web API сервиса
- gRPC-Gateway в `gateway-service` для REST-маршрутов
- Zap для логирования
- Prometheus для метрик
- OpenTelemetry + Jaeger для трассировки
- golang-migrate для миграций
- Docker Compose

## Переменные окружения

Полный пример: [`.env.example`](./.env.example).

| Переменная | Назначение | По умолчанию |
|---|---|---|
| `TELEGRAM_BOT_TOKEN` | Токен BotFather | обязательно |
| `TELEGRAM_MODE` | `webhook` или `polling` | `webhook` |
| `TELEGRAM_WEBHOOK_URL` | Публичный HTTPS URL webhook | `TELEGRAM_PUBLIC_URL + /telegram/webhook` |
| `TELEGRAM_WEBHOOK_SECRET` | Секрет заголовка `X-Telegram-Bot-Api-Secret-Token` | обязательно для webhook |
| `TELEGRAM_PUBLIC_URL` | Публичный URL сервиса для Web App | `http://localhost:8082` |
| `TELEGRAM_WEBAPP_PATH` | Путь Web App | `/telegram/webapp` |
| `TELEGRAM_WEBAPP_MAX_AGE` | Максимальный возраст `initData` | `24h` |
| `TELEGRAM_WEBAPP_TOKEN_TTL` | Срок JWT для Web App | `15m` |
| `TELEGRAM_SERVICE_TOKEN_TTL` | Срок JWT для вызовов Gateway | `5m` |
| `TELEGRAM_LINK_CODE_TTL` | Срок кода привязки | `10m` |
| `TELEGRAM_DIALOG_TTL` | TTL состояния диалога | `30m` |
| `TELEGRAM_UPDATE_DEDUP_TTL` | TTL ключа обработанного update | `24h` |
| `TELEGRAM_EVENT_DEDUP_TTL` | TTL ключа обработанного Kafka-события | `168h` |
| `TELEGRAM_HTTP_TIMEOUT` | Таймаут вызовов Gateway | `120s` |
| `TELEGRAM_NOTIFICATION_RETRIES` | Число попыток доставки | `3` |
| `TELEGRAM_HTTP_ADDR` | HTTP-адрес webhook/Web App/health | `:8082` |
| `TELEGRAM_GRPC_ADDR` | gRPC-адрес сервиса | `:50062` |
| `TELEGRAM_METRICS_ADDR` | Адрес Prometheus | `:9100` |
| `GATEWAY_URL` | Gateway внутри Docker-сети | `http://gateway:8081` |
| `GATEWAY_PUBLIC_URL` | Gateway для Web App | `http://localhost:8081` |
| `POSTGRES_*` | PostgreSQL сервиса | — |

Локально сервис читает стандартные `POSTGRES_*` (как остальные микросервисы). В корневом `docker-compose.yaml` эти значения пробрасываются из `TELEGRAM_POSTGRES_*`.
| `REDIS_ADDR` | Redis | `localhost:6379` |
| `KAFKA_BROKERS` | Kafka brokers | `localhost:9092` |
| `KAFKA_TOPIC` | Топик событий | `finance-events` |
| `KAFKA_CONSUMER_GROUP` | Consumer group | `telegram-bot-service` |
| `JWT_SECRET` | Общий секрет платформы | обязательно |
| `JWT_DURATION` | Срок пользовательского JWT | `24h` |
| `LOGGER_LEVEL`, `LOGGER_FOLDER`, `TIME_ZONE` | Логирование и часовой пояс | — |

Секреты не коммитятся. `.env` исключён из Git.

## Настройка Telegram BotFather

1. Откройте [@BotFather](https://t.me/BotFather).
2. Выполните `/newbot`, задайте имя и username.
3. Скопируйте токен в `TELEGRAM_BOT_TOKEN`.
4. Для Web App задайте публичный HTTPS-адрес сервиса в `TELEGRAM_PUBLIC_URL`.
5. При необходимости выполните `/setcommands` и добавьте:
   - `start` — привязка и главное меню;
   - `menu` — главное меню;
   - `balance` — финансовый обзор;
   - `settings` — настройки;
   - `unlink` — отвязать Telegram;
   - `help` — справка.

## Настройка Webhook

Webhook устанавливается автоматически при старте в режиме `webhook`:

```text
POST https://api.telegram.org/bot<TOKEN>/setWebhook
url=https://<TELEGRAM_PUBLIC_URL>/telegram/webhook
secret_token=<TELEGRAM_WEBHOOK_SECRET>
allowed_updates=["message","callback_query"]
```

Требования:

- публичный HTTPS-адрес;
- заголовок `X-Telegram-Bot-Api-Secret-Token` проверяется в постоянное время;
- один и тот же update не обрабатывается дважды благодаря Redis `SET NX`;
- входящий update ставится в ограниченную очередь и обрабатывается worker-пулом;
- при переполнении очереди возвращается `503`, чтобы Telegram повторил доставку.

Для локальной разработки без HTTPS используйте `TELEGRAM_MODE=polling`. В этом режиме webhook не устанавливается, а обновления забираются Long Polling.

Проверка webhook:

```bash
curl "https://api.telegram.org/bot<TOKEN>/getWebhookInfo"
```

## Привязка аккаунта

```text
Веб-приложение
  1. Пользователь входит в аккаунт.
  2. Открывает раздел «Telegram-бот» в боковом меню.
  3. Нажимает «Подключить Telegram».
  4. Gateway вызывает TelegramService.CreateLinkCode.
  5. Backend возвращает одноразовый код (TTL 10 минут).

Telegram
  6. Пользователь отправляет /start КОД.
  7. Бот ищет код по SHA-256 хэшу.
  8. Проверяет TTL, признак использования и отсутствие конфликтов.
  9. Атомарно помечает код использованным и создаёт привязку.
 10. Создаёт настройки уведомлений по умолчанию.
```

Правила безопасности:

- в веб-приложении добавлен раздел «Telegram-бот» с кнопками подключения и отвязки;
- пароль в Telegram не запрашивается и не хранится;
- код хранится только в виде SHA-256;
- код одноразовый и ограничен по времени;
- один Telegram ID нельзя привязать к двум аккаунтам;
- к одному аккаунту нельзя привязать два Telegram ID;
- `/unlink` удаляет привязку и настройки уведомлений;
- web-отзыв выполняется через `DELETE /api/v1/telegram/link`;
- JWT выпускается только после успешного разрешения привязки;
- токены и `initData` не пишутся в логи.

## Интеграция с Gateway

Бот использует REST API Gateway и передаёт `Authorization: Bearer <JWT>`. JWT подписывается общим `JWT_SECRET` и содержит `user_id` привязанного аккаунта.

Используемые маршруты:

| Метод | Маршрут | Назначение |
|---|---|---|
| GET | `/api/v1/finance/dashboard` | обзор |
| GET | `/api/v1/finance/transactions` | список операций с фильтрами |
| POST | `/api/v1/finance/transaction` | создание операции |
| GET/PUT/DELETE | `/api/v1/finance/transaction/{id}` | операция |
| GET | `/api/v1/finance/categories` | категории |
| GET | `/api/v1/finance/analytics` | аналитика |
| GET/POST/PUT/DELETE | `/api/v1/finance/budget(s)(/{id})` | бюджеты |
| GET | `/api/v1/finance/export/{csv,pdf,json,txt}` | экспорт |
| GET | `/api/v1/finance/export/file?key=...` | скачивание файла |
| GET/POST/PUT/DELETE/PATCH | `/api/v1/shopping...` | списки покупок |
| POST | `/api/v1/ai/plan` | AI-план |
| GET/DELETE | `/api/v1/ai/plans...` | история планов |
| GET | `/api/v1/currency/{base}` | курсы |
| POST | `/api/v1/currency/convert` | конвертация |
| GET | `/api/v1/user/{id}` | профиль |

Новый маршрут `/api/v1/finance/export/file` добавлен в Gateway, потому что существующий маршрут скачивания по `{key}` не принимает ключи S3, содержащие `/`. Новый маршрут принимает ключ query-параметром и возвращает файл потоком.

## Интеграция с Kafka

Consumer group: `telegram-bot-service`. Топик: `finance-events`.

| Событие | Producer | Действие бота |
|---|---|---|
| `budget.exceeded` | finance-service | уведомление о превышении бюджета |
| `ai.plan_created` | ai-service | уведомление о завершении AI-планирования |
| `transaction.created` | finance-service | уведомление о новой операции |
| `user.deleted` | users-service | удаление привязки и настроек |

`ai.plan_created` публикуется `ai-service` после успешного сохранения плана. Формат события:

```json
{
  "plan_id": 12,
  "user_id": 42,
  "request_id": "uuid",
  "total_cost": 1234.5,
  "mode": "self_drive",
  "created_at": "2026-01-01T12:00:00Z"
}
```

Каждое событие идемпотентно: ключ `telegram:event:<type>:<id>` хранится в Redis. При ошибке доставки ключ освобождается, а сообщение не коммитится, что позволяет повторить обработку.

## Telegram Web App

Web App доступен по адресу `TELEGRAM_PUBLIC_URL + TELEGRAM_WEBAPP_PATH`. Он:

- инициализирует `Telegram.WebApp`;
- отправляет `initData` на `POST /telegram/webapp/session`;
- получает короткоживущий JWT после серверной проверки подписи;
- загружает AI-план через Gateway;
- показывает магазины, цены, доставку, рекомендацию и маршрут на Leaflet/OpenStreetMap.

Проверка `initData`:

1. `hash` извлекается из query string.
2. Формируется data-check-string из отсортированных пар `key=value`.
3. `secret_key = HMAC_SHA256("WebAppData", bot_token)`.
4. `expected = HMAC_SHA256(secret_key, data_check_string)`.
5. Сравнение выполняется в постоянное время.
6. `auth_date` проверяется на свежесть.
7. Только после этого `user.id` используется для поиска привязки.

Web App не доверяет данным клиента до проверки подписи и не принимает `user_id` из запроса.

## Локальный запуск

1. Создайте окружение:

   ```bash
   cp telegram-bot-service/.env.example telegram-bot-service/.env
   ```

2. Заполните `TELEGRAM_BOT_TOKEN`, `JWT_SECRET`, параметры PostgreSQL, Redis и Kafka.

3. Для локальной разработки установите `TELEGRAM_MODE=polling`.

4. Примените миграции:

   ```bash
   migrate -path telegram-bot-service/migrations \
     -database "postgres://telegram:telegram@localhost:5432/telegram_bot?sslmode=disable" up
   ```

5. Запустите сервис:

   ```bash
   cd telegram-bot-service
   make run
   ```

   `make` загружает `.env`. При прямом `go run ./cmd/telegram/main.go` переменные нужно экспортировать в оболочку.

Сервис требует доступный `gateway-service` и остальные сервисы платформы. Без них бот запускается, но пользовательские операции будут возвращать сообщение о недоступности.

## Запуск через Docker Compose

Из корня репозитория:

```bash
cp env.example .env
# заполните TELEGRAM_BOT_TOKEN, TELEGRAM_WEBHOOK_URL, TELEGRAM_WEBHOOK_SECRET
docker compose up -d --build gateway telegram-bot
```

Проверка:

```bash
docker compose ps telegram-bot
curl http://localhost:8082/healthz
curl http://localhost:9100/metrics
```

Сервис подключён к сети `finance-net`, использует общие `finance-redis` и `kafka`, собственный `telegram-postgres`, публикует порты `8082`, `50062`, `9100`. Остальные контейнеры не изменяются.

Для внешнего webhook пробросьте `8082` на публичный HTTPS-домен через reverse proxy. Telegram требует HTTPS.

## Тесты

```bash
cd telegram-bot-service
go test ./...
go vet ./...
go build ./...
```

Интеграционные тесты включаются переменными окружения и иначе пропускаются. Используйте отдельную тестовую базу и Redis: тесты очищают таблицы `telegram.*` и создают временные ключи.

```bash
export TEST_POSTGRES_DSN="postgres://telegram:telegram@localhost:5432/telegram_bot_test?sslmode=disable"
export TEST_REDIS_ADDR="localhost:6379"
export TEST_KAFKA_BROKERS="localhost:9092"
go test ./... -count=1
```

Покрытые сценарии:

- привязка аккаунта, TTL, повторное использование кода, конфликт привязок;
- состояния диалогов;
- создание и редактирование операции с описанием;
- валидация сумм и дат;
- подтверждение удаления;
- уведомления, повторные попытки, отключение, `403` от Telegram;
- обработка Kafka-событий и идемпотентность;
- проверка Telegram `initData` (валидная, изменённая, истёкшая, будущая);
- Gateway-клиент: заголовок авторизации, фильтры, коды ошибок, скачивание файла, координаты AI;
- webhook: секрет, очередь, некорректный метод;
- Web App session: выпуск JWT, отказ для непривязанного пользователя;
- PostgreSQL-репозитории, Redis-состояния и Kafka-consumer (при наличии окружения).

## Безопасность

- Секреты не хранятся в коде и не коммитятся.
- В логах нет токенов, паролей и `initData`.
- Telegram ID → user ID разрешается только через собственную БД.
- JWT имеет короткий срок и выпускается только после привязки.
- Коды привязки хранятся в виде хэша и одноразовые.
- Webhook защищён секретным заголовком.
- Повторные update и Kafka-события отсекаются идемпотентно.
- Все пользовательские ресурсы проверяются на принадлежность в доменных сервисах.

## Известные ограничения

- Реальная доставка сообщений зависит от доступности Telegram Bot API; в тестах используется mock.
- Web App требует HTTPS; локально кнопка открывается как обычная ссылка.
- Позиции внутри списка покупок не поддерживаются существующим API.
- Уведомления по обновлению и удалению транзакций не отправляются.
- Для webhook нужен публичный HTTPS-домен и reverse proxy; сервис не терминирует TLS на порту 8082.
