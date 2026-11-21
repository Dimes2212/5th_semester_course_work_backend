# CampusLost Backend

Backend Android-приложения CampusLost — сервиса потерянных и найденных вещей для студентов и сотрудников МИРЭА.

Документ одновременно является техническим заданием и планом реализации. Если код, миграции и этот документ расходятся, расхождение необходимо явно обсудить и исправить: новые возможности не должны появляться в проекте случайно.

## 1. Состояние проекта

Функциональный MVP backend реализован:

- HTTP API на Go и `chi`;
- регистрация, JWT access-токены и ротируемые refresh-сессии;
- профиль и удаление аккаунта;
- справочники, объявления, фильтрация и до трёх фотографий;
- приватное S3-совместимое хранилище Backblaze B2 с подписанными URL;
- диалоги, история сообщений и WebSocket-событие `message_created`;
- административные маршруты, блокировка пользователей и модерация объявлений;
- PostgreSQL 18.6, миграции `golang-migrate` и health check.

Точный машиночитаемый контракт API находится в [`docs/openapi.yaml`](docs/openapi.yaml).

## 2. Функциональные требования

Пользователь должен иметь возможность:

1. Зарегистрироваться с любым корректным адресом электронной почты.
2. Войти в приложение и обновлять сессию без повторного ввода пароля.
3. Выйти из текущей сессии.
4. Получить и отредактировать свой профиль.
5. Получить доступные категории и места.
6. Создать объявление о потерянной или найденной вещи.
7. Прикрепить к объявлению не более трёх фотографий.
8. Просматривать ленту активных объявлений с пагинацией.
9. Искать объявления и фильтровать их по типу, категории и месту.
10. Открыть отдельную карточку объявления.
11. Просматривать собственные объявления, включая завершённые.
12. Редактировать и удалять собственные объявления.
13. Подтвердить возврат вещи владельцу.
14. Создать диалог по объявлению.
15. Просматривать список своих диалогов и историю сообщений.
16. Отправлять сообщения.

### Пока не входят в MVP

- избранные объявления и лайки;
- жалобы пользователей;
- push-уведомления;
- восстановление пароля;
- подтверждение email кодом;
- отдельный интерфейс админ-панели;
- автоматическая отправка системных Android-уведомлений.

Эти возможности добавляются отдельными миграциями и задачами после завершения основного API.

## 3. Технологии

- Go;
- `chi` — HTTP-роутер;
- `pgx/v5` и `pgxpool` — работа с PostgreSQL;
- PostgreSQL 18.6;
- `golang-migrate` — версионирование схемы;
- JWT access-токены;
- случайные refresh-токены с хранением только их хэшей;
- bcrypt — хэширование паролей;
- `coder/websocket` — WebSocket-доставка событий;
- `go-chi/httprate` — in-memory rate limit;
- Docker Compose — локальное окружение;
- Backblaze B2 через S3-совместимый API — хранение фотографий.

Не используются ORM, микросервисы, Kafka и другие инструменты, которые не нужны для текущего масштаба приложения.

## 4. Бизнес-правила

### 4.1 Пользователи и сессии

- Email перед сохранением приводится к нижнему регистру.
- Регистрация разрешена с любым корректным email независимо от почтового домена.
- Пароль и его подтверждение никогда не сохраняются в исходном виде.
- В PostgreSQL хранится bcrypt-хэш пароля.
- Каждый пользователь имеет роль `user` или `admin`; публичная регистрация всегда создаёт роль `user`.
- Администратор предсоздаётся миграцией; публичного API создания администратора или изменения роли нет.
- Access-токен имеет короткий срок жизни и не хранится в БД.
- Refresh-токен является криптографически случайной строкой.
- В БД хранится только SHA-256-хэш refresh-токена.
- При обновлении сессии refresh-токен ротируется: старый отзывается, новый сохраняется.
- Выход отзывает refresh-сессию.
- Каждый защищённый запрос дополнительно проверяет, что пользователь не удалён и не заблокирован, поэтому блокировка немедленно инвалидирует и уже выпущенные access-токены.

Рекомендуемые сроки:

- access token — 15 минут;
- refresh token — 30 дней.

### 4.2 Объявления

Тип объявления `kind`:

- `lost` — вещь потеряна;
- `found` — вещь найдена.

Статус объявления `status`:

- `active` — отображается в общей ленте;
- `resolved` — вещь возвращена;
- `archived` — объявление закрыто автором.

Обычная лента возвращает только записи:

```sql
status = 'active' AND deleted_at IS NULL
```

После возврата вещь не удаляется физически. Объявление получает статус `resolved`, исчезает из общей ленты, но остаётся доступным в истории автора и для будущей админ-панели.

Удаление объявления пользователем является мягким: заполняется `deleted_at`.

### 4.3 Места

Основное место выбирается из справочника `locations`. Пользователь не вводит произвольное название корпуса.

Поле `location_details` используется для уточнения:

```text
аудитория 314
около гардероба
второй этаж
```

### 4.4 Фотографии

- Объявление может содержать от нуля до трёх фотографий.
- В PostgreSQL хранится только `storage_key`, а не содержимое файла.
- `position` может принимать значения `1`, `2` или `3`.
- У одного объявления не может быть двух фотографий с одинаковой позицией.
- Допустимые форматы MVP: JPEG, PNG и WebP.
- Максимальный размер одного файла: 8 МБ.
- Сервер обязан проверять реальный MIME-тип, а не только расширение файла.

Ограничения `CHECK (position BETWEEN 1 AND 3)` и `UNIQUE (listing_id, position)` гарантируют максимум три позиции на объявление.

### 4.5 Диалоги

- Диалог создаётся по конкретному объявлению.
- Инициатор не может начать диалог по собственному объявлению.
- Для пары `объявление + инициатор` существует не более одного диалога.
- Повторный запрос на создание должен вернуть существующий диалог.
- Участниками являются инициатор и автор объявления.
- Отправлять и читать сообщения могут только участники диалога.
- Длина сообщения — от 1 до 2000 символов после удаления пробелов по краям.

## 5. Модель данных

Актуальная схема создаётся миграцией [`migrations/000001_initial_schema.up.sql`](migrations/000001_initial_schema.up.sql).

```mermaid
erDiagram
    USERS ||--o{ REFRESH_SESSIONS : "имеет сессии"
    USERS ||--o{ LISTINGS : "создаёт"
    USERS o|--o{ LISTINGS : "получает возвращённую вещь"
    CATEGORIES ||--o{ LISTINGS : "классифицирует"
    LOCATIONS ||--o{ LISTINGS : "указывает место"
    LISTINGS ||--o{ LISTING_PHOTOS : "содержит до 3 фото"
    LISTINGS ||--o{ CONVERSATIONS : "имеет диалоги"
    USERS ||--o{ CONVERSATIONS : "инициирует"
    CONVERSATIONS ||--o{ MESSAGES : "содержит"
    USERS ||--o{ MESSAGES : "отправляет"

    USERS {
        uuid id PK
        varchar full_name
        varchar email UK
        text password_hash
        varchar role
        text avatar_key
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    REFRESH_SESSIONS {
        uuid id PK
        uuid user_id FK
        text token_hash UK
        varchar device_name
        timestamptz expires_at
        timestamptz created_at
        timestamptz revoked_at
    }

    CATEGORIES {
        smallint id PK
        varchar slug UK
        varchar name UK
        smallint position
        boolean is_active
    }

    LOCATIONS {
        smallint id PK
        varchar name UK
        varchar address
        smallint position
        boolean is_active
    }

    LISTINGS {
        uuid id PK
        uuid author_id FK
        smallint category_id FK
        smallint location_id FK
        varchar location_details
        varchar kind
        varchar title
        text description
        timestamptz event_at
        varchar status
        uuid resolved_with_user_id FK
        timestamptz resolved_at
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    LISTING_PHOTOS {
        uuid id PK
        uuid listing_id FK
        text storage_key
        smallint position
        timestamptz created_at
    }

    CONVERSATIONS {
        uuid id PK
        uuid listing_id FK
        uuid initiator_id FK
        timestamptz created_at
        timestamptz updated_at
    }

    MESSAGES {
        uuid id PK
        uuid conversation_id FK
        uuid sender_id FK
        text text
        timestamptz created_at
    }
```

### Начальные категории

- Электроника;
- Документы;
- Одежда;
- Ключи;
- Компьютерная периферия;
- Другое.

### Начальные места

- Стромынка 20;
- Проспект Вернадского 78;
- Проспект Вернадского 86;
- Спорткомплекс;
- Коворкинг;
- Столовая;
- Другое.

Таблица `schema_migrations` является технической и создаётся `golang-migrate`.

## 6. Архитектура Go-приложения

Файлы создаются постепенно, по мере реализации возможностей. На первом этапе проект содержит только точку входа, инфраструктуру приложения и пакет авторизации.

```text
cmd/
└── api/
    └── main.go

internal/
├── admin/
├── apihelper/
├── auth/
├── chat/
├── health/
├── listing/
├── lookup/
├── ratelimit/
├── realtime/
├── requestctx/
├── s3/
└── user/

migrations/
docs/
└── openapi.yaml
```

### Ответственность пакетов

| Пакет | Ответственность |
|---|---|
| `cmd/api` | Чтение окружения, подключение PostgreSQL, сборка зависимостей, регистрация маршрутов и lifecycle HTTP-сервера. |
| `auth/service.go` | Сценарии регистрации, входа, обновления токенов и выхода. |
| `auth/service_helpers.go` | Валидация, генерация токенов и преобразования сервисного слоя. |
| `auth/handler.go` | HTTP-запросы и ответы авторизации. |
| `apihelper` | Общая работа с JSON и единым форматом HTTP-ошибок. |
| `auth/handler_helpers.go` | Преобразование ошибок и ответов авторизации. |
| `auth/repository.go` | Работа с `users` и `refresh_sessions` через PostgreSQL. |
| `listing` | Объявления, фильтрация и фотографии. |
| `chat` | Диалоги, история и отправка сообщений. |
| `realtime` | Hub подключений и WebSocket-handler. |
| `admin` | Блокировка пользователей и модерация объявлений. |
| `health` | Проверка PostgreSQL для оркестратора. |
| `ratelimit` | Ограничение частоты чувствительных запросов. |

Файлы внутри feature-пакета разделяются по handler/service/repository только там, где это помогает сохранить понятные границы ответственности.

Направление зависимостей:

```text
handler -> service.go (бизнес-логика) -> repository -> PostgreSQL
```

Правила:

- handler не выполняет SQL;
- repository не формирует HTTP-ответы;
- бизнес-логика в `service.go` не зависит от `http.ResponseWriter` или `*http.Request`;
- бизнес-проверки выполняются вне handler;
- SQL всегда параметризован;
- зависимости создаются в `cmd/api/main.go`, а не через глобальные переменные;
- интерфейс вводится только на реальной границе или для тестирования, а не для каждой структуры автоматически;
- предметные пакеты не должны импортировать друг друга по кругу.

## 7. HTTP API

Базовый префикс:

```text
/api/v1
```

Формат времени — RFC 3339 с часовым поясом:

```text
2026-11-04T14:30:00+03:00
```

Идентификаторы передаются строками UUID.

### 7.1 Формат успешного ответа

```json
{
  "data": {}
}
```

Для коллекции:

```json
{
  "data": [],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 47,
    "total_pages": 3
  }
}
```

### 7.2 Формат ошибки

```json
{
  "error": "validation_error"
}
```

Основные коды:

| HTTP | `code` | Значение |
|---:|---|---|
| 400 | `invalid_request` | Некорректный JSON, query или multipart-запрос. |
| 400 | `validation_error` | Поля не прошли валидацию. |
| 401 | `unauthorized` | Нет access-токена или он недействителен. |
| 403 | `forbidden` | Пользователь авторизован, но действие ему запрещено. |
| 404 | `not_found` | Ресурс не найден или недоступен пользователю. |
| 409 | `conflict` | Email уже зарегистрирован или нарушена уникальность. |
| 413 | `payload_too_large` | Превышен допустимый размер файла или тела. |
| 415 | `unsupported_media_type` | Недопустимый тип фотографии. |
| 429 | `rate_limit_exceeded` | Превышен лимит запросов. |
| 500 | `internal_error` | Внутренняя ошибка без раскрытия деталей. |
| 503 | `service_unavailable` | PostgreSQL или зависимость временно недоступны. |

### 7.3 Авторизация

Защищённые запросы используют заголовок:

```http
Authorization: Bearer <access-token>
```

#### Регистрация

```http
POST /api/v1/auth/register
Content-Type: application/json
```

```json
{
  "full_name": "Иван Иванов",
  "email": "ivanov@example.com",
  "password": "strong-password",
  "password_confirmation": "strong-password",
  "device_name": "Samsung Galaxy S24"
}
```

Проверки:

- имя после `TrimSpace` не пустое и не длиннее 100 символов;
- email имеет корректный формат и не длиннее 255 символов;
- email ещё не зарегистрирован;
- пароль содержит минимум 8 символов и не длиннее 72 байт для bcrypt;
- пароль совпадает с подтверждением.

Ответ `201 Created` содержит пользователя, access-токен и refresh-токен. Поле `password_hash` никогда не попадает в API.

#### Вход

```http
POST /api/v1/auth/login
```

```json
{
  "email": "ivanov@example.com",
  "password": "strong-password",
  "device_name": "Samsung Galaxy S24"
}
```

Для неверного email и неверного пароля возвращается одинаковая ошибка, чтобы не раскрывать существование аккаунта.

#### Обновление токенов

```http
POST /api/v1/auth/refresh
```

```json
{
  "refresh_token": "random-token"
}
```

При успехе старый refresh-токен отзывается и возвращается новая пара токенов.

#### Выход

```http
POST /api/v1/auth/logout
Authorization: Bearer <access-token>
```

```json
{
  "refresh_token": "random-token"
}
```

Успешный ответ: `204 No Content`.

Отдельный endpoint «проверить сессию» не нужен. Android проверяет access-токен запросом `GET /me`; после `401` пытается выполнить refresh.

### 7.4 Профиль

#### Получить текущего пользователя

```http
GET /api/v1/me
```

#### Изменить профиль

```http
PATCH /api/v1/me
```

```json
{
  "full_name": "Иван Петров"
}
```

#### Аватар

```http
PUT    /api/v1/me/avatar
DELETE /api/v1/me/avatar
```

Загрузка использует `multipart/form-data`. Старая фотография удаляется только после успешного сохранения новой и обновления БД.

#### Собственные объявления

```http
GET /api/v1/me/listings?page=1&per_page=20&status=active
```

Допустимые фильтры `status`: `active`, `resolved`, `archived`. Если фильтр отсутствует, возвращаются все неудалённые объявления пользователя.

### 7.5 Справочники

```http
GET /api/v1/categories
GET /api/v1/locations
```

Возвращаются только записи с `is_active = true`, отсортированные по `position`.

### 7.6 Объявления

#### Лента, поиск и фильтры

```http
GET /api/v1/listings
```

Query-параметры:

| Параметр | Тип | По умолчанию | Назначение |
|---|---|---:|---|
| `page` | integer | `1` | Номер страницы. |
| `per_page` | integer | `20` | Размер страницы, максимум `100`. |
| `q` | string | — | Поиск по названию и описанию. |
| `kind` | `lost` или `found` | — | Тип объявления. |
| `category_id` | smallint | — | Категория. |
| `location_id` | smallint | — | Основное место. |

Endpoint без параметров возвращает первую страницу общей ленты. Отдельного endpoint «получить вообще все объявления без ограничений» не существует.

Сортировка MVP:

```sql
ORDER BY created_at DESC, id DESC
```

Первая реализация поиска может использовать `ILIKE`. Полнотекстовый поиск или `pg_trgm` добавляются отдельной оптимизацией после измерения производительности.

#### Получить объявление

```http
GET /api/v1/listings/{listingID}
```

Активное объявление доступно всем авторизованным пользователям. Неактивное объявление в MVP доступно только его автору.

#### Создать объявление

```http
POST /api/v1/listings
Content-Type: multipart/form-data
```

Поля:

| Поле | Обязательное | Ограничение |
|---|---|---|
| `kind` | да | `lost` или `found` |
| `title` | да | 1–120 символов после `TrimSpace` |
| `description` | нет | до 2000 символов |
| `category_id` | да | активная категория |
| `location_id` | да | активное место |
| `location_details` | нет | до 150 символов |
| `event_at` | да | RFC 3339 |
| `photos` | нет | 0–3 файла |

Создание объявления и записей `listing_photos` выполняется транзакционно. Если сохранение БД завершилось ошибкой, уже загруженные файлы должны быть удалены.

#### Изменить поля объявления

```http
PATCH /api/v1/listings/{listingID}
Content-Type: application/json
```

Менять объявление может только автор. `author_id`, `status`, `resolved_at` и `resolved_with_user_id` этим endpoint не изменяются.

#### Управление фотографиями

```http
POST   /api/v1/listings/{listingID}/photos
DELETE /api/v1/listings/{listingID}/photos/{photoID}
```

Добавление запрещено, если после операции станет больше трёх фотографий.

#### Удалить объявление

```http
DELETE /api/v1/listings/{listingID}
```

Только автор. Заполняет `deleted_at`, ответ `200 OK` содержит `id` и `deleted_at`.

#### Подтвердить возврат

```http
POST /api/v1/listings/{listingID}/resolve
```

Только автор активного объявления. Тело запроса не требуется. Операция устанавливает `status = resolved` и `resolved_at = now()`.

### 7.7 Чаты

#### Создать или получить диалог

```http
POST /api/v1/listings/{listingID}/conversations
```

Если диалог уже существует, возвращается существующий объект. Автор объявления не может создать диалог с самим собой.

#### Список диалогов

```http
GET /api/v1/me/conversations
```

Элемент списка содержит ID диалога, ID и заголовок объявления, инициатора, время создания/обновления и текст последнего сообщения либо `null`.

#### История сообщений

```http
GET /api/v1/conversations/{conversationID}/messages
```

#### Отправить сообщение

```http
POST /api/v1/conversations/{conversationID}/messages
```

```json
{
  "text": "Привет! Кажется, это мои наушники."
}
```

Ответ: `201 Created` с сохранённым сообщением.

### 7.8 WebSocket

```http
GET /api/v1/ws
Authorization: Bearer <access-token>
```

Соединение используется только для событий от сервера. Сообщения создаются REST-запросом. При новом сообщении онлайн-получатель получает:

```json
{
  "type": "message_created",
  "data": {
    "id": "message-uuid",
    "conversation_id": "conversation-uuid",
    "sender_id": "user-uuid",
    "text": "Привет!",
    "created_at": "2026-11-21T12:00:00Z"
  }
}
```

PostgreSQL остаётся источником истины: после переподключения клиент перечитывает диалоги и историю через REST.

### 7.9 Администрирование

Все маршруты требуют access-токен с ролью `admin`:

```http
GET    /api/v1/admin/users
PATCH  /api/v1/admin/users/{userID}/status?blocked=true|false
GET    /api/v1/admin/listings
DELETE /api/v1/admin/listings/{listingID}
PATCH  /api/v1/admin/listings/{listingID}/restore
```

Блокировка отзывает активные refresh-сессии и немедленно запрещает использование access-токенов. Удаление объявления является мягким; фотографии, диалоги и сообщения сохраняются.

### 7.10 Системный endpoint

```http
GET /health
```

`200 OK`, если HTTP-сервер работает и PostgreSQL отвечает на `Ping`. При недоступной БД возвращает `503 Service Unavailable`.

## 8. Конфигурация

Планируемые переменные окружения API:

```dotenv
APP_ENV=local
HTTP_ADDRESS=:8080
TLS_CERT_FILE=
TLS_KEY_FILE=
DATABASE_URL=postgres://campuslost:campuslost@localhost:5436/campuslost?sslmode=disable
JWT_SECRET=replace-with-long-random-secret
JWT_ISSUER=campuslost
ACCESS_TOKEN_TTL=15m
REFRESH_TOKEN_TTL=720h

S3_ENDPOINT=s3.<region>.backblazeb2.com
S3_REGION=<region>
S3_BUCKET=<bucket-name>
S3_ACCESS_KEY_ID=
S3_SECRET_ACCESS_KEY=
S3_USE_SSL=true
```

Требования:

- обязательные переменные проверяются при запуске;
- приложение завершается с понятной ошибкой при некорректной конфигурации;
- секреты не коммитятся;
- `.env` игнорируется Git;
- `.env.example` содержит только безопасные локальные значения и имена переменных.
- в `APP_ENV=production` приложение не запускается без `DATABASE_URL` и `JWT_SECRET`;
- если заданы `TLS_CERT_FILE` и `TLS_KEY_FILE`, сервер запускает HTTPS напрямую; иначе TLS завершается на доверенном reverse proxy.

## 9. Локальный запуск PostgreSQL

Требования: Docker и Docker Compose.

```bash
make db-up
make migrate-up
```

Проверить таблицы:

```bash
make db-tables
```

Открыть `psql`:

```bash
make db-shell
```

Посмотреть логи:

```bash
make db-logs
```

Остановить контейнер без удаления данных:

```bash
make db-down
```

Локальная строка подключения с хоста:

```text
postgres://campuslost:campuslost@localhost:5436/campuslost?sslmode=disable
```

Внутри Docker-сети PostgreSQL доступен как `postgres:5432`.

Откатить последнюю миграцию:

```bash
make migrate-down
```

Команда отката изменяет схему и может удалить данные. Её следует выполнять осознанно.

Создать локальную резервную копию в формате `pg_dump --format=custom`:

```bash
make db-backup
```

Файлы сохраняются в игнорируемом каталоге `backups/`. Восстановление выполняется явно:

```bash
make db-restore BACKUP_FILE=backups/campuslost_YYYYMMDD_HHMMSS.dump
```

Команда восстановления перезаписывает объекты текущей базы и должна запускаться осознанно.

## 10. Безопасность

Обязательные требования MVP:

- bcrypt для паролей;
- криптографически случайные refresh-токены;
- хранение только хэша refresh-токена;
- ротация refresh-токена;
- проверка `exp`, подписи и типа JWT;
- параметризованные SQL-запросы;
- проверка авторства объявлений;
- проверка участия в диалоге;
- проверка активности пользователя в БД на каждом защищённом запросе;
- проверка роли для административных маршрутов;
- лимит размера HTTP body и фотографий;
- проверка MIME-типа файлов;
- случайные имена сохраняемых файлов;
- запрет `..` и пользовательских путей в `storage_key`;
- HTTP timeouts;
- graceful shutdown;
- recovery middleware;
- request ID;
- структурированные логи без паролей и токенов;
- единый формат ошибок без текста внутренних SQL-ошибок;
- rate limit на регистрацию, вход, refresh и отправку сообщений;
- HTTPS и `wss` в production.

Для локальной разработки допустим HTTP. Release-сборка Android должна обращаться только к HTTPS API.

## 11. Транзакции

Транзакция обязательна, когда одна операция меняет несколько связанных записей:

- регистрация пользователя и создание refresh-сессии;
- ротация refresh-токена;
- создание объявления и записей фотографий;
- подтверждение возврата вещи;
- создание диалога при конкурентных запросах.

Файловая система не участвует в транзакции PostgreSQL, поэтому service обязан выполнять компенсацию: удалять новый файл, если операция в БД не завершилась.

## 12. Тестирование

### Unit-тесты service

Проверяют бизнес-правила без HTTP:

- регистрация с некорректным форматом email;
- несовпадающие пароли;
- редактирование чужого объявления;
- четвёртая фотография;
- диалог с самим собой;
- отправка сообщения посторонним пользователем;
- повторное завершение объявления.

### Handler-тесты

Через `httptest` проверяют:

- HTTP-статусы;
- декодирование JSON и query;
- обязательность авторизации;
- формат успешных ответов и ошибок.

### Интеграционные тесты repository

Запускаются с реальным PostgreSQL и проверяют:

- SQL-запросы;
- ограничения схемы;
- пагинацию и фильтры;
- конкурентное создание одного диалога;
- транзакционный откат.

## 13. Порядок реализации и коммиты

Каждый коммит должен собираться, проходить тесты и представлять один понятный этап.

1. `chore: add PostgreSQL environment and initial schema`
   - Docker Compose;
   - Makefile;
   - миграция и справочники;
   - документация схемы.

2. `chore: initialize Go API application`
   - `go.mod`;
   - config;
   - `pgxpool`;
   - chi router;
   - graceful shutdown;
   - `/health`.

3. `feat: implement user registration`
   - bcrypt;
   - нормализация и проверка email;
   - создание пользователя и refresh-сессии;
   - тесты.

4. `feat: implement login and session refresh`
   - login;
   - access JWT;
   - refresh rotation;
   - logout;
   - auth middleware.

5. `feat: implement user profile`
   - `GET /me`;
   - `PATCH /me`;
   - аватар.

6. `feat: expose categories and locations`
   - справочники;
   - repository и handler tests.

7. `feat: implement listing creation`
   - локальное хранилище;
   - создание объявления;
   - до трёх фотографий;
   - транзакция и компенсация файлов.

8. `feat: implement listing feed and search`
   - пагинация;
   - фильтры;
   - поиск;
   - карточка объявления.

9. `feat: implement listing management`
   - собственные объявления;
   - редактирование;
   - фотографии;
   - мягкое удаление;
   - завершение.

10. `feat: implement conversations`
    - создание диалога;
    - список диалогов;
    - проверки участников.

11. `feat: implement chat messages`
    - история;
    - отправка.

12. `docs: add OpenAPI specification`
    - полный контракт для Android-клиента.

13. `ci: add automated checks`
    - форматирование;
    - `go vet`;
    - `go test ./...`.

WebSocket и административные маршруты реализованы. Production deployment и Android push-уведомления остаются отдельными этапами.

## 14. Критерии готовности endpoint

Endpoint считается готовым, если:

- маршрут зарегистрирован;
- запрос и ответ описаны;
- входные данные валидируются;
- права доступа проверяются;
- бизнес-логика находится в service;
- SQL находится в repository и параметризован;
- ошибки преобразуются в единый формат;
- написаны положительные и отрицательные тесты;
- `go test ./...` проходит;
- код отформатирован `gofmt`;
- секреты и персональные данные не выводятся в лог.

## 15. Принятые решения

- Архитектура feature-first: пакеты делятся по сущностям.
- Используется PostgreSQL 18.6 на локальном порту `5436`.
- ORM не используется.
- Избранного в MVP нет.
- В системе предусмотрены только роли `user` и `admin`, без отдельной роли модератора.
- Место выбирается из справочника.
- Категория «Другое» существует.
- На объявление разрешено не более трёх фотографий.
- Возвращённое объявление скрывается из ленты, но сохраняется в БД.
- Сообщения сохраняются через REST, а WebSocket доставляет событие `message_created` онлайн-получателю.
- Администратор может блокировать обычных пользователей, просматривать, скрывать и восстанавливать объявления.
- Системные Android push-уведомления не входят в текущий MVP.
