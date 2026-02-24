# 1. Общая информация

## 1.1 Назначение

Микросервис обработки обращений пользователей (тикетов), включающий:

- управление категориями
- создание и обработку тикетов
- чат внутри тикета (WebSocket)
- назначение тикетов сотрудникам поддержки
- управление статусами
- realtime-события

---

# 2. Роли и права доступа (RBAC)

## 2.1 Роли

- `user`
- `support`
- `admin`

---

## 2.2 Матрица доступа

| Действие             | user     | support         | admin |
| -------------------- | -------- | --------------- | ----- |
| Создать тикет        | ✅        | ❌               | ❌     |
| Смотреть свои тикеты | ✅        | ❌               | ❌     |
| Смотреть open тикеты | ❌        | ✅               | ✅     |
| Назначить тикет      | ❌        | ✅ (только себе) | ✅     |
| Менять статус        | ❌        | ✅ (свои)        | ✅     |
| Писать сообщения     | ✅ (свои) | ✅ (назначенные) | ✅     |
| CRUD категорий       | ❌        | ❌               | ✅     |

---

# 3. Функциональные требования

---

# 3.1 Категории

### Доступ: admin

### Операции:

- Создание
- Удаление
- Получение

---

# 3.2 Тикеты

## Создание

- Только user
- Статус: `open`
- assigned_to = NULL

---

## Назначение

- support может назначить только на себя
- admin может назначить на любого support
- при назначении:
    - статус → `in_progress`

---

## Закрытие

- support (назначенный)
- admin
- user (только свои)

Статус → `closed`

---

## Получение списка
### user
- только свои
### support
- open
- assigned_to = self
### admin
- все

---
# 3.3 Сообщения

- Привязаны к тикету
- Отправитель определяется по JWT
- После сохранения:
    - рассылается WebSocket-событие


---

# 4. Сущности

## 4.1 Category

```go
type Category struct {
    ID      int64
    Name    string
    Enabled bool
}
```

---
## 4.2 Ticket

```go
type Ticket struct {
    ID         int
    CategoryID int
    CreatorID  int
    AssignedTo *int
    Status     string
    Subject    string
    Source     string
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
```

TicketStatus:  {open, in_progress, closed}
TicketSource: {web, mobile, service}

---

## 4.3 Message

```go
type Message struct {
    ID         int64  `db:"id"`
    TicketID   int    `db:"ticket_id"`
    SenderID   int    `db:"sender_id"`
    SenderType string `db:"sender_type"`
    Content    string `db:"content"`
    CreatedAt  string `db:"created_at"`
}
```
### SenderType
- user
- support
- admin
- bot (???)

---

# 5. База данных

---

## 5.1 categories
```sql
id bigserial primary key
name text not null
enabled boolean not null default true
destination text not null,
created_at timestamp not null
updated_at timestamp not null
```
---

## 5.2 tickets
```sql
id bigserial primary key
category_id bigint references categories(id)
creator_id int not null
assigned_to int
status text not null
subject text not null
source text not null
created_at timestamp not null
updated_at timestamp not null
```
---

## 5.3 messages
```sql
id bigserial primary key,
ticket_id integer not null references tickets(id),
sender_id integer not null,
sender_type text not null,
content text not null,
created_at timestamp not null
```
---

# 6. REST API
### Список всех endpoints

```
POST  /categories
GET   /categories
PUT   /categories/{id}
PATCH /categories/{id}/disable
PATCH /categories/{id}/enable

POST  /tickets
GET   /tickets
GET   /tickets/{id}
PATCH  /tickets/{id}/assign
PATCH /tickets/{id}/status

POST  /tickets/{id}/messages
```
---

## 6.1 Categories

### POST /categories

Создание новой категории.
Доступ: admin

Request:
```json
{
  "name": "Payments",
  "destination": "user"
}
```

Логика
- Проверить роль admin
- Создать запись
- enabled = true
- created_at = now

Response:
```json
{
  "id": 1,
  "name": "Payments",
  "enabled": true,
  "created_at": "2026-02-17T10:00:00Z"
}
```
---

### GET /categories

Получить список всех категорий.
Доступ: admin, user (active only)

Response:
```json
[
  {
    "id": 1,
    "name": "Payments",
    "enabled": true
  },
  {
    "id": 2,
    "name": "Technical",
    "enabled": false
  }
]
```
---

### PUT /categories/{id}

Обновление категории.
Доступ: admin

Request:
```json
{
  "name": "Billing",
  "enabled": true
}
```

Response:
```json
{
  "id": 1,
  "name": "Billing",
  "enabled": true
}
```
---
## 6.2 Tickets

### POST /tickets

Создание тикета пользователем.
Доступ: user

Request:

```json
{
  "category_id": 1,
  "message": "I can't complete payment",
  "source": "mobile",
  "subject": "App error"
}
```

Response:
```json
{
  "ticket_id": 10
}
```

Логика:
1. Проверить, что категория существует и enabled=true
2. Создать тикет:
    - status = open
    - creator_id = JWT.sub
3. Создать первое сообщение
---

### GET /tickets

Получить список тикетов.
Доступ: user (только свои), support (open + assigned_to=self), admin (все)

Query параметры:
`?status=open`
`?status=in_progress`
`?status=closed`

Response:
```json
[
  {
    "id": 10,
    "category_id": 1,
    "creator_id": "123",
    "assigned_to": "456",
    "status": "in_progress",
    "created_at": "2026-02-17T10:10:00Z"
  }
]
```
---

### GET /tickets/{id}

Получить конкретный тикет + историю сообщений.
Доступ: user (только свой), support (только assigned), admin (любой)

Response:
```json
{
  "id": 10,
  "category_id": 1,
  "creator_id": "123",
  "assigned_to": "456",
  "status": "in_progress",
  "messages": [
    {
      "id": 1,
      "sender_id": "123",
      "sender_type": "user",
      "content": "I can't complete payment",
      "created_at": "2026-02-17T10:10:00Z"
    },
    {
      "id": 2,
      "sender_id": "456",
      "sender_type": "support",
      "content": "Please send screenshot",
      "created_at": "2026-02-17T10:12:00Z"
    }
  ]
}

```
---

### PATCH /tickets/{id}/assign

Назначить тикет.
Доступ: support (только себе), admin (может указать assigned_to)

Request:
```json
{
  "assigned_to": "456"
}
```

Логика
- если саппорт то проверяем что assigned_to == userID
- assigned_to = userID
- status → in_progress

Response
```json
{
  "ticket_id": 10,
  "assigned_to": "456",
  "status": "in_progress"
}
```
---

### PATCH /tickets/{id}/status

Изменить статус тикета.
Доступ: user (только свой), support (assigned), admin

Request:
```json
{
  "status": "closed"
}
```

Response:
```json
{
  "ticket_id": 10,
  "status": "closed"
}
```
---

## 6.3 Messages

### POST /tickets/{id}/messages

Добавить сообщение в тикет.
Доступ: user (только свой тикет), support (только назначенный), admin (любой)

Request:
```json
{
  "content": "Here is screenshot"
}
```

Response:
```json
{
  "message_id": 5,
  "created_at": "2026-02-17T10:15:00Z"
}
```

Логика:
1. Проверка доступа
2. Создание записи в messages
3. WebSocket broadcast
___

# 7. WebSocket архитектура

---

## Endpoint

GET /ws

*JWT обязателен.

### Входящие события от клиента

#### join_ticket
```json
{
  "type": "join_ticket",
  "ticket_id": 10
}
```

#### leave_ticket
```json
{
  "type": "leave_ticket"
}
```

---
### Исходящие события от сервера

#### new_message
```json
{
  "type": "new_message",
  "data": {
    "ticket_id": 10,
    "message_id": 5,
    "sender_type": "support",
    "content": "Please check again"
  }
}
```

#### ticket_assigned
```json
{
  "type": "ticket_assigned",
  "data": {
    "ticket_id": 10,
    "assigned_to": "456"
  }
}

```

#### ticket_status_changed
```json
{
  "type": "ticket_status_changed",
  "data": {
    "ticket_id": 10,
    "status": "closed"
  }
}
```
---

## Hub
- хранит подключения
- управляет регистрацией
- рассылает события

---