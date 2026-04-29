
# Bank API — RESTful банковский сервис на Go

Учебный проект, реализующий REST API для банковского сервиса с поддержкой счетов, карт, кредитов, аналитики и интеграций с Центральным банком РФ и SMTP-уведомлениями.

## Основные возможности

- **Пользователи**: регистрация, аутентификация (JWT).
- **Счета**: создание, пополнение, снятие, просмотр списка счетов.
- **Карты**: выпуск виртуальной карты (алгоритм Луна), просмотр (с расшифровкой), оплата.
- **Переводы**: между счетами с транзакционной целостностью.
- **Кредиты**: оформление с аннуитетным расчётом, график платежей, автоматическое списание, штрафы за просрочку.
- **Аналитика**: доходы/расходы за месяц, кредитная нагрузка, прогноз баланса на N дней.
- **Интеграции**:
  - ЦБ РФ: получение ключевой ставки через SOAP (KeyRate).
  - SMTP: отправка email-уведомлений о платежах и штрафах.
- **Безопасность**:
  - Пароли: bcrypt.
  - Номера и сроки карт: PGP + HMAC-SHA256.
  - CVV: bcrypt.
  - JWT-аутентификация и проверка прав доступа.
- **Планировщик** (scheduler): обработка просроченных кредитных платежей каждые 12 часов.

## Технологии и библиотеки

- Go 1.23+
- gorilla/mux (роутер)
- PostgreSQL 17 + lib/pq
- golang-jwt/jwt/v5
- logrus (логирование)
- bcrypt (пароли, CVV)
- PGP: github.com/ProtonMail/go-crypto/openpgp
- HMAC-SHA256
- go-mail/mail v2 (SMTP)
- beevik/etree (парсинг XML)
- shopspring/decimal (точная работа с деньгами)
- go-playground/validator (валидация входных данных)

## Структура проекта
bank-api/
├── cmd/main.go
├── internal/
│ ├── config/config.go
│ ├── models/
│ ├── repository/
│ ├── service/
│ ├── handler/
│ ├── middleware/
│ ├── integration/
│ ├── encryption/
│ ├── scheduler/
│ └── router/
├── migrations/001_init.sql
├── tests/integration_test.go
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
└── README.md



## Установка и запуск

### Предварительные требования

- Go 1.23 или новее
- PostgreSQL 17 с расширением `pgcrypto`
- GnuPG для генерации PGP-ключей (опционально)

### 1. Клонирование репозитория

```bash
git clone github.com/ingenjor/bank-api
cd bank-api
2. Настройка переменных окружения
Скопируйте .env.example в .env и заполните значения. Обязательные переменные:


DATABASE_URL=postgres://bank_user:bank_pass@localhost:5432/bank_db?sslmode=disable
JWT_SECRET=super-secret-jwt-key
3. Генерация PGP-ключей (опционально)
bash
mkdir -p keys
gpg --quick-generate-key "Bank App <bank@example.com>" rsa4096 encr never
gpg --armor --export bank@example.com > keys/public.asc
gpg --armor --export-secret-keys bank@example.com > keys/private.asc
Если ключи не сгенерированы, функционал карт будет недоступен, но всё остальное будет работать.

4. Настройка базы данных
Создайте базу данных PostgreSQL и выполните миграцию:

bash
psql -U bank_user -d bank_db -f migrations/001_init.sql
Или используйте Docker:

bash
docker run --name bank-postgres -e POSTGRES_USER=bank_user -e POSTGRES_PASSWORD=bank_pass -e POSTGRES_DB=bank_db -p 5432:5432 -d postgres:17
5. Установка зависимостей и запуск
bash
go mod tidy
go run cmd/main.go
Сервер запустится на порту, указанном в SERVER_PORT (по умолчанию 8080).

API Эндпоинты
Публичные
POST /register – регистрация (body: {"username","email","password"})

POST /login – вход, возвращает JWT

Защищённые (требуют Authorization: Bearer <token>)
Все маршруты под префиксом /api.

Счета
POST /api/accounts – создать счёт

GET /api/accounts – список счетов пользователя

POST /api/accounts/{id}/deposit – пополнить (body: {"amount": 1000})

POST /api/accounts/{id}/withdraw – снять (body: {"amount": 500})

Карты
POST /api/cards – выпустить карту (body: {"account_id":"uuid"})

GET /api/cards – все карты пользователя

GET /api/cards/{id} – информация о карте

POST /api/cards/payment – оплата картой (body: {"card_id":"uuid","amount":150})

Переводы
POST /api/transfer – перевод между счетами (body: {"from","to","amount"})

Кредиты
POST /api/credits – оформить кредит (body: {"amount":100000,"term_months":12})

GET /api/credits/{creditId}/schedule – график платежей

Аналитика
GET /api/analytics – сводка за месяц

GET /api/accounts/{accountId}/predict?days=90 – прогноз баланса

Тестирование
Unit-тесты
bash
go test ./internal/... -v
Интеграционные тесты (требуют тестовую БД)
Добавьте в .env переменную TEST_DATABASE_URL и выполните:

bash
go test -tags=integration -v ./tests/
Безопасность
Пароли хешируются bcrypt.

Номера и сроки карт шифруются PGP с открытым ключом и подписываются HMAC-SHA256.

CVV хешируется bcrypt.

Все защищённые эндпоинты проверяют JWT и принадлежность ресурса пользователю.

Планировщик
Каждые 12 часов обрабатываются просроченные кредитные платежи: при достаточном балансе – списание, иначе – штраф 10% (однократно).

Лицензия
Проект учебный.
