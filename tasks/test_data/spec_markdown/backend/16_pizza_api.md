# Техническое задание

## Название проекта

Pizza Ordering API

## Цель

Разработать REST API для оформления заказов пиццы с авторизацией пользователей через JWT.

## Технологический стек

### Backend

* Go 1.26.3
* net/http
* PostgreSQL 17.9
* Redis 8.6.3
* pgx
* go-redis
* golang-jwt

### Инфраструктура

* Docker
* Docker Compose

### Архитектура

Hexagonal Architecture (Ports & Adapters)

---

# Функциональные требования

## Авторизация

### Регистрация пользователя

**Endpoint**

POST /auth/register

**Request**

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Validation**

* email обязателен
* email должен быть уникальным
* password обязателен
* password не менее 8 символов

**Response**

```json
{
  "message": "user registered"
}
```

---

### Авторизация пользователя

**Endpoint**

POST /auth/login

**Request**

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response**

```json
{
  "access_token": "...",
  "refresh_token": "..."
}
```

---

### JWT

Использовать два типа токенов:

#### Access Token

* срок жизни: 15 минут
* используется для доступа к защищенным эндпоинтам

#### Refresh Token

* срок жизни: 7 дней
* хранится в Redis

---

## Заказы пиццы

### Создание заказа

**Endpoint**

POST /pizza/create/order

**Authorization**

Bearer Access Token

**Request**

```json
{
  "pizza_name": "Pepperoni",
  "size": "large",
  "quantity": 2
}
```

**Validation**

* pizza_name обязателен
* size обязателен
* допустимые значения:

  * small
  * medium
  * large
* quantity > 0

**Response**

```json
{
  "id": "uuid",
  "status": "created"
}
```

---

### Получение списка заказов

**Endpoint**

GET /pizza/orders

**Authorization**

Bearer Access Token

**Response**

```json
[
  {
    "id": "uuid",
    "pizza_name": "Pepperoni",
    "size": "large",
    "quantity": 2,
    "created_at": "2026-01-01T12:00:00Z"
  }
]
```

Пользователь должен видеть только свои заказы.

# Redis

Использовать Redis для хранения refresh токенов.

---

# Архитектура проекта

```text
cmd/
└── app/

internal/
├── domain/
│   ├── user.go
│   └── order.go
│
├── ports/
│   ├── repository.go
│   ├── service.go
│   └── jwt.go
│
├── application/
│   ├── auth/
│   └── pizza/
│
├── adapters/
│   ├── http/
│   │   ├── handlers/
│   │   └── middleware/
│   │
│   ├── postgres/
│   │   └── repositories/
│   │
│   ├── redis/
│   │   └── repositories/
│   │
│   └── jwt/
│
pkg/
configs/
migrations/
```

---

# Docker Compose

Необходимо подготовить docker-compose.yml, содержащий:

### Сервисы

#### app

* golang:1.26.3-alpine

#### postgres

* postgres:17.9-alpine3.23

Параметры:

```env
POSTGRES_DB=pizza
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
```

Порт:

```yaml
5432:5432
```

---

#### redis

* redis:8.6.3-alpine3.23

Порт:

```yaml
6379:6379
```

---

# Переменные окружения

```env
HTTP_PORT=8080

POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_DB=pizza
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres

REDIS_ADDR=redis:6379

JWT_SECRET=super-secret-key

ACCESS_TOKEN_TTL=15m
REFRESH_TOKEN_TTL=168h
```

---

# Нефункциональные требования

* Использовать context.Context во всех слоях.
* Использовать pgx для работы с PostgreSQL.
* Использовать go-redis для работы с Redis.
* Пароли хранить только в виде bcrypt-хеша.
* Возвращать корректные HTTP-коды ошибок.
* JSON-ответы должны быть единообразны.
* Код должен соответствовать принципам Hexagonal Architecture.
* Без использования сторонних HTTP-фреймворков (Gin, Echo, Fiber и т.д.).
* Использовать только net/http.

---

# Критерии приемки

Проект считается выполненным, если:

1. Пользователь может зарегистрироваться.
2. Пользователь может авторизоваться.
3. Генерируются access и refresh токены.
4. Refresh токен сохраняется в Redis.
5. Авторизованный пользователь может создать заказ пиццы.
6. Авторизованный пользователь может получить список своих заказов.
7. PostgreSQL и Redis запускаются через Docker Compose.
8. Приложение запускается через Docker Compose.
9. Код организован согласно Hexagonal Architecture.
10. Все эндпоинты корректно работают через Postman или curl.
11. Проверка на правильность загрузки переменных окружения.