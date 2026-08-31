# Finance Microservices

Платформа для учёта личных финансов на микросервисной архитектуре (Go). Клиенты обращаются к системе только через gateway-service по REST/JSON, а он уже проксирует запросы во внутренние gRPC-вызовы к доменным сервисам. Между собой сервисы общаются синхронно через gRPC и асинхронно через Kafka.

Репозиторий — Go multi-module монорепо, каждый сервис имеет свой go.mod, свой Dockerfile и запускается отдельным процессом.


## Стек технологий
Язык и основная экосистема

Go 1.26.4

Коммуникация

gRPC — синхронное взаимодействие между сервисами

REST/JSON (через gRPC-Gateway) — для внешних клиентов

Apache Kafka (через segmentio/kafka-go) — асинхронные события

Базы данных и хранилища

PostgreSQL 15 (три изолированные БД: auth, users, finance)

Redis 7 — кэш и read-модели

MinIO (S3-совместимое хранилище) — для экспортированных файлов

Инфраструктура и оркестрация

Docker и Docker Compose

GitHub Actions — CI/CD

Наблюдаемость

Prometheus — метрики

Jaeger — распределённая трассировка (OpenTelemetry)

Zap — структурированное логирование

Kafka UI — просмотр событий

Документация и тестирование

Swagger / OpenAPI (через grpc-gateway)

Testify + mock — юнит-тесты

Внешние API

Frankfurter API — курсы валют

Хостинг

Timeweb Cloud (РФ, Санкт-Петербург)

SSH - безопастность

## Хостинг
Проект запущен на сервере Timeweb Cloud в Санкт-Петербурге. Конфигурация: 8 vCPU, 8 ГБ RAM, 100 ГБ NVMe. Публичный IP: 185.125.200.63.

Доступ по SSH через ключи. Вся инфраструктура поднимается через Docker Compose на одном хосте. Все шесть микросервисов работают стабильно.

Проект оплачен до 05.09.2026 после его снимат с хостинга

Публичные эндпоинты:

API Gateway — http://185.125.200.63:8081

Swagger — http://185.125.200.63:8081/swagger/

Kafka UI — порт 8080

MinIO Console — порт 9001

Jaeger — порт 16686

Prometheus — порт 9090

Настроен CI/CD через GitHub Actions. При пуше в master запускаются проверки и тесты, затем код автоматически обновляется на сервере через docker compose up -d --build.

Безопасность: JWT для API, секреты хранятся в .env на сервере. Логи пишутся в ./logs, метрики в Prometheus, трейсы в Jaeger.

## Архитектура

Клиент ходит только в gateway-service по HTTP. Gateway валидирует JWT и проксирует запросы по gRPC в пять доменных сервисов: auth, users, finance, admin, currency. У auth, users и finance есть собственная база PostgreSQL (database-per-service, у каждого своя изолированная база). admin и currency своей базы не имеют — они читают данные через gRPC-вызовы к другим сервисам и слушают Kafka, а всё что им нужно хранить (агрегаты, кэш, курсы валют) держат в Redis.

Все сервисы завязаны на общий Kafka-топик finance-events — через него сервисы узнают об изменениях в других доменах, не завязываясь друг на друга напрямую. Redis общий для всех сервисов (кэш профилей, списков, курсов, метрик). MinIO используется только в finance-service — туда складываются экспортированные отчёты.

Структура кода у каждого Go-сервиса одинаковая:
- cmd/<service> — точка входа main.go
- config — конфиг из переменных окружения через envconfig
- internal/core — инфраструктурные адаптеры: пул postgres, redis, kafka, grpc-клиенты/серверы, метрики, логгер, трейсинг,domain
- internal/features/<domain>/service — бизнес-логика
- internal/features/<domain>/repository/postgres и /redis — доступ к данным
- internal/features/<domain>/transport/grpc и /kafka — gRPC-хендлеры и Kafka-продюсеры/консюмеры
- proto — сгенерированный protobuf/gRPC код
- migrations — SQL-миграции через golang-migrate
- pkg — общие утилиты: errors, logger, pagination, grpcutil, types,middelware

## Сервисы

### gateway-service

Единая точка входа для внешних клиентов. HTTP-сервер на базе grpc-gateway, который автоматически строит REST-роуты /api/v1/... из аннотаций в proto-файлах и проксирует их как gRPC-вызовы к auth, users, finance, admin, currency.

Что делает: проверяет JWT (Bearer-заголовок или cookie token) на всех маршрутах кроме /healthz, /metrics, /swagger/*, /api/v1/auth/register, /api/v1/auth/login, /api/v1/auth/logout; прокидывает заголовок Authorization в метаданные gRPC-вызова, чтобы доменные сервисы сами проверяли токен и роль; отдаёт отдельный HTTP-эндпоинт для скачивания экспортированных файлов из S3 (/api/v1/finance/export/download?key=...); раздаёт Swagger UI и сгенерированную OpenAPI-спеку; отдаёт метрики Prometheus на /metrics; есть /healthz для healthcheck. Цепочка middleware: CORS → Auth (JWT) → RequestID → Logger → OpenTelemetry Trace → Panic recovery.

### auth-service

Отвечает за аутентификацию и хранение учётных данных — email и bcrypt-хэш пароля, отдельно от профиля пользователя.

gRPC-методы: Login (проверка пароля, выдача JWT), Register (создаёт credential у себя, потом вызывает users-service.CreateProfile для создания профиля, выдаёт JWT), Logout.

Своя PostgreSQL-база с таблицей credentials, у записи есть статус pending/active/failed. Redis используется для кэша/сессий. Публикует и потребляет события через общий топик finance-events. JWT генерируется и валидируется через golang-jwt/jwt.

### users-service

Хранилище профилей пользователей — email, имя, телефон, роль admin/user. Источник истины о пользователях.

gRPC-методы: GetUser, GetUserByEmail, ListUsers, PatchUser, DeleteUser, CreateProfile (создание профиля, вызывается из auth-service.Register), UpdateRole (назначение/снятие роли администратора), AdminExists (проверка что в системе есть хотя бы один админ), GetMetrics, а также MarkDeleting, FinalizeDelete, RestoreUser — это шаги саги удаления пользователя, подробнее ниже.

Своя PostgreSQL-база с миграциями. Кэш профиля и кэш списка пользователей в Redis. При создании/удалении пользователя публикует события user.created и user.deleted в Kafka, чтобы admin и currency держали свой read-model в актуальном состоянии.

### finance-service

Основной сервис финансового учёта — транзакции (доходы/расходы), дашборд, категории, экспорт отчётов.

gRPC-методы: CreateTransaction, GetTransaction, GetTransactions, UpdateTransaction, DeleteTransaction, DeleteUserTransaction (массовое удаление всех транзакций пользователя, это шаг саги удаления пользователя, вызывается из admin-service), GetDashboard (сводная статистика по доходам/расходам), GetCategories, GetMetrics, ExportJSON/ExportCSV/ExportTXT/ExportPDF (генерируют отчёт в нужном формате и заливают файл в MinIO по ключу вида exports/user_<id>/export_<date>.<ext>), DownloadExport (забирает файл обратно из S3 чтобы отдать клиенту через gateway).

Своя PostgreSQL-база с транзакциями пользователей. Redis-кэш. S3/MinIO через aws-sdk-go-v2/service/s3 — используется исключительно для хранения экспортированных отчётов, транзакционные данные там не хранятся. PDF генерируется через gofpdf. При создании/изменении/удалении транзакции публикует события transaction.created, transaction.updated, transaction.deleted в Kafka, чтобы admin (агрегированная статистика) и currency (конвертация в USD) обновляли свои read-модели.

### admin-service

Административная панель — управление пользователями, статистика по системе, удаление пользователей вместе со всеми их данными в других сервисах.

gRPC-методы: GetUsers, GetUser (проксирует данные из users-service по gRPC), DeleteUser (запускает сагу удаления пользователя, координируя users-service и finance-service), UpdateUserRole, GetMetrics (агрегированная статистика, строится по событиям Kafka и кэшируется в Redis).

Своей PostgreSQL-базы нет — использует Redis как хранилище агрегированных метрик и кэша. Является Kafka-консюмером: слушает transaction.created, transaction.updated, transaction.deleted, user.created, user.deleted из finance-events и обновляет свои агрегаты — счётчики, дашборд администратора. Это классический CQRS/read-model на событиях. Является и Kafka-продюсером: публикует admin.metrics и user.transactions.deleted. Реализует оркестрацию саги удаления пользователя с явной компенсацией при ошибке.

### currency-service

Конвертация валют и приведение сумм транзакций к USD.

gRPC-методы: GetRates (текущие курсы валют), Convert и ConvertBatch (конвертация суммы/пакета сумм между валютами), GetTransactionUSD (сумма транзакции в пересчёте на USD).

Своей PostgreSQL-базы нет, курсы валют и рассчитанные суммы хранятся в Redis. Актуальные курсы берутся из внешнего REST API Frankfurter (https://api.frankfurter.app/latest) — бесплатный сервис курсов валют ЕЦБ. Является Kafka-консюмером: слушает transaction.created и transaction.deleted из finance-events, чтобы пересчитывать и кэшировать сумму транзакции в USD сразу при её появлении или удалении, без похода в finance-service на каждый запрос.

## Межсервисное взаимодействие

### Паттерн Saga

В системе используется оркестрируемая saga — распределённая транзакция, затрагивающая несколько сервисов с собственными базами данных, где согласованность достигается последовательностью локальных транзакций и компенсирующими действиями при сбое, вместо распределённого 2PC.

Saga удаления пользователя, оркестратор — admin-service, метод DeleteUser:

Сначала admin-service вызывает users-service.MarkDeleting(id) — помечает пользователя как "в процессе удаления", это мягкая блокировка чтобы не потерять данные при частичном сбое. Дальше admin-service вызывает finance-service.DeleteUserTransactions(id) — удаляет все транзакции пользователя. Если этот шаг падает, выполняется компенсация: вызывается users-service.RestoreUser(id), чтобы откатить пользователя из состояния "удаляется" обратно в нормальное, и сага прерывается с ошибкой — состояние системы остаётся консистентным. Если удаление транзакций прошло успешно, admin-service вызывает users-service.FinalizeDelete(id), который окончательно удаляет пользователя. Вся сага логируется целиком (Starting DeleteUser saga / saga completed successfully), что позволяет отследить весь процесс в логах и в трейсинге через Jaeger.

Специально под этот сценарий в users-service выделены три отдельных gRPC-метода жизненного цикла — MarkDeleting, RestoreUser, FinalizeDelete, а не один общий DeleteUser — это и есть техническая реализация шагов саги с возможностью компенсации.

Кроме этого регистрация пользователя через auth-service.Register тоже по сути мини-сага с двумя локальными транзакциями в разных базах: сначала создаётся запись credential в auth-service со статусом pending, потом вызывается users-service.CreateProfile, и в зависимости от результата статус credential переводится в active либо failed.

### События Kafka

Единый топик finance-events используется как шина событийного взаимодействия между сервисами, реализовано через segmentio/kafka-go. Каждый сервис поднимает продюсера и/или консюмера в зависимости от роли.

transaction.created публикует finance-service, слушают admin-service и currency-service — обновление статистики и пересчёт суммы в USD. transaction.updated публикует finance-service, слушает admin-service — обновление статистики. transaction.deleted публикует finance-service, слушают admin-service и currency-service — обновление статистики и инвалидация USD-кэша. user.created и user.deleted публикует users-service, слушает admin-service — обновление счётчиков пользователей. user.transactions.deleted публикует finance-service как сигнал о завершении шага саги удаления транзакций. admin.metrics публикует admin-service после пересчёта агрегированных метрик.

Просмотр и отладка сообщений в топике доступны через веб-интерфейс Kafka UI.

## Инфраструктура

Полное описание в docker-compose.yaml, поднимается одной командой docker compose up -d.

finance-postgres, auth-postgres, users-postgres — по одной изолированной базе postgres:15-alpine на каждый домен, database-per-service.

finance-redis на redis:7-alpine — общий кэш и хранилище для всех сервисов: профили, списки, курсы валют, метрики. kafka на confluentinc/cp-kafka — брокер событий в режиме KRaft без Zookeeper, топик finance-events. kafka-ui на provectuslabs/kafka-ui — веб-интерфейс для просмотра топиков и сообщений Kafka, http://localhost:8080. minio на minio/minio — S3-совместимое объектное хранилище для экспортированных финансовых отчётов в JSON/CSV/TXT/PDF, S3 API на порту 9000, веб-консоль на http://localhost:9001.

jaeger на jaegertracing/all-in-one — распределённый трейсинг через OpenTelemetry/OTLP, UI на http://localhost:16686. prometheus на prom/prometheus — сбор метрик со всех сервисов, http://localhost:9090. auth-postgres-migrate, users-postgres-migrate, finance-postgres-migrate на migrate/migrate — одноразовые job-контейнеры для прогона SQL-миграций. auth, users, finance, admin, currency, gateway — доменные сервисы приложения, каждый собирается из своего Dockerfile в build/<service>/.

Все сервисы объединены в одну bridge-сеть Docker finance-net.

## Наблюдаемость

Метрики: каждый сервис отдаёт метрики Prometheus на своём порту через /metrics, Prometheus их собирает по конфигу prometheus.yml. Смотреть агрегированные метрики: http://localhost:9090.

Трейсинг: все сервисы инструментированы через OpenTelemetry (gRPC- и HTTP-инструментация), трейсы уходят по OTLP/gRPC в Jaeger. Смотреть трейсы запросов между сервисами: http://localhost:16686.

Логирование: структурированные логи через zap, пишутся в файлы в папку ./logs, уровень настраивается через LOGGER_LEVEL.

Kafka UI — просмотр сообщений и consumer group'ов в топике finance-events: http://localhost:8080.

MinIO Console — просмотр загруженных файлов экспорта отчётов: http://localhost:9001, S3 API на http://localhost:9000.

Swagger UI — интерактивная документация REST API gateway: http://localhost:8081/swagger/.

## Внешние библиотеки

Общие для всех или большинства Go-сервисов:

google.golang.org/grpc — gRPC-сервер и клиент, основной протокол межсервисного взаимодействия.

google.golang.org/protobuf — работа с Protocol Buffers.

grpc-ecosystem/grpc-gateway/v2 — генерация REST-прослойки JSON↔gRPC из proto-аннотаций, используется в gateway. 

jackc/pgx/v5 — драйвер и пул соединений PostgreSQL.

redis/go-redis/v9 — клиент Redis.

segmentio/kafka-go — продюсер и консюмер Apache Kafka.

golang-jwt/jwt/v5 — генерация и валидация JWT.

golang.org/x/crypto (bcrypt) — хэширование паролей.

google/uuid — генерация UUID.

kelseyhightower/envconfig — маппинг переменных окружения в конфиг-структуры.

uber-go/zap — структурированное логирование.

prometheus/client_golang — экспорт метрик Prometheus.

open-telemetry/opentelemetry-go (otel, otel/sdk, otlptracegrpc, otelgrpc, otelhttp) — распределённый трейсинг и экспорт в Jaeger по OTLP.

stretchr/testify и golang/mock — юнит-тестирование и моки. golang-migrate/migrate — версионирование и прогон SQL-миграций, используется как отдельный сервис в docker-compose.

Специфичные для отдельных сервисов:

aws/aws-sdk-go-v2 (service/s3) — клиент S3 API, используется finance-service для загрузки и чтения файлов экспорта в MinIO.

jung-kurt/gofpdf — генерация PDF-отчётов в finance-service.

go-playground/validator/v10 — валидация входящих HTTP-запросов в gateway.

swaggo/http-swagger и swaggo/swag — генерация и раздача Swagger UI и OpenAPI-спеки в gateway. Frankfurter API (https://api.frankfurter.app/latest) — внешний REST-источник курсов валют для currency-service.

## Порты сервисов

gateway-service — HTTP на 8081 (REST API и Swagger), gRPC не используется.

auth-service — gRPC на 50051, метрики на 9091.

users-service — gRPC на 50052, метрики на 9096.

finance-service — gRPC на 50053, метрики на 9093.

admin-service — gRPC на 50054, метрики на 9094.

currency-service — gRPC на 50055, метрики на 9095.

Инфраструктура: PostgreSQL (auth/users/finance) доступны только внутри docker-сети без внешнего проброса. Redis на 6379. Kafka на 9092. Kafka UI на 8080. MinIO S3 API на 9000, MinIO Console на 9001. Jaeger UI на 16686, Jaeger OTLP gRPC на 4317. Prometheus на 9090.

## Переменные окружения

Полный список — в env.example.
Основные группы:

POSTGRES_*,

AUTH_POSTGRES_*,

USERS_POSTGRES_*,

FINANCE_POSTGRES_* — доступы к соответствующим базам.

JWT_SECRET и JWT_DURATION — секрет и срок жизни JWT.

REDIS_ADDR — адрес Redis.

KAFKA_BROKERS — адрес брокера Kafka.

S3_ENDPOINT_URL, S3_REGION, S3_BUCKET, S3_ACCESS_KEY, S3_SECRET_KEY — доступ к MinIO/S3. MINIO_ROOT_USER и MINIO_ROOT_PASSWORD — учётные данные root MinIO.

TLS_ENABLED, TLS_CERT_FILE, TLS_KEY_FILE, TLS_CA_FILE — опциональный TLS для gRPC-соединений между сервисами.

LOGGER_LEVEL и LOGGER_FOLDER — настройки логирования.

TIME_ZONE — таймзона процессов.

## Запуск проекта

Скопировать env.example в .env и заполнить значения, затем docker compose up -d --build.

Основные команды из Makefile: make up поднимает всю систему, make down останавливает, make migrate-up-all прогоняет миграции для auth/users/finance, make services-logs показывает логи доменных сервисов, make kafka-ui открывает Kafka UI, make proto-gen-all перегенерирует protobuf/gRPC код из .proto, make swagger-gen перегенерирует Swagger из аннотаций.

После запуска: 

REST API и Swagger на http://localhost:8081/swagger/

Kafka UI на http://localhost:8080

MinIO Console на http://localhost:9001

Jaeger на http://localhost:16686

Prometheus на http://localhost:9090
