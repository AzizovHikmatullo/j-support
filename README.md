# **Документация проекта J-Support**
## 1. Общее описание проекта

**J-Support** - это бэкенд-система для обработки обращений клиентов через разные каналы (Web-виджет, Telegram, мобильное приложение).

Основные возможности:
- Создание и управление тикетами
- Многоуровневые сценарии бота (дерево шагов с условиями)
- Реал-тайм обмен сообщениями через WebSocket
- Разделение прав (user / support / admin)
- Автоматическое закрытие неактивных тикетов
- Полный аудит действий (Activity Log)
- Поддержка нескольких каналов общения

## 2. Архитектура

- **Язык:** Go 1.25
- **Фреймворк:** Gin
- **База данных:** PostgreSQL
- **WebSocket:** Gorilla WebSocket + собственный Hub
- **Очереди/планировщик:** gocron
- **Миграции:** `migrations/`

## 3. Авторизация и идентификация

Система использует **два типа авторизации** одновременно:

### 3.1. Для поддержки и администраторов
- **Header:** `Authorization: Bearer <JWT>`
- JWT содержит: `userID` и `role` (`support` или `admin`)

### 3.2. Для клиентов (пользователей)
Используется middleware `ChannelIdentityMiddleware`. Клиент может авторизоваться тремя способами:

| Способ                  | Header                    | ChannelType   | Role  |
|-------------------------|---------------------------|---------------|-------|
| Web-виджет              | `X-Session-Token`         | `web`         | user  |
| Telegram                | `X-Telegram-ID`           | `telegram`    | user  |
| Мобильное приложение    | `Authorization: Bearer`   | `app`         | user  |

**Важно:** Все клиентские запросы проходят через `ChannelIdentityMiddleware`.

## 4. Основные сущности

### 4.1. Contact (Контакт)
- Один контакт может быть привязан к `user_id`, `external_id` (telegram/web) и телефону.
- Создаётся автоматически при первом обращении.

### 4.2. Category (Категория)
- Имеет `destination` (`user`, `driver` и т.д.)
- Может быть включена/выключена.
- К каждой категории может быть привязан активный сценарий бота.

### 4.3. Ticket (Тикет)
Статусы:
- `pending` - создан, бот работает
- `open` - открыт для общения
- `in_progress` - назначен сотруднику
- `closed` - закрыт

### 4.4. Message (Сообщение)
- Может содержать кнопки (только от бота)

### 4.5. Scenario + Step (Сценарии бота)
- Сценарий привязывается к категории.
- Шаги образуют **дерево**.
- Каждый шаг может иметь `condition` (условие) или быть **default** (без условия).
- Поддерживается только один default-переход с одного шага.

### 4.6. ActivityLog
Фиксирует все действия:
- `created`, `status_changed`, `assigned`, `message_sent`, `rated`

## 5. Основные сценарии работы

### 5.1. Создание тикета клиентом
1. Клиент отправляет `POST /tickets`
2. Создаётся тикет в статусе `pending`
3. Система проверяет наличие активного сценария для категории
4. Если сценарий есть → запускается бот (первый вопрос + кнопки)
5. Если сценария нет → тикет сразу переходит в `open`

### 5.2. Работа бота (Scenario)
- Бот работает **только** пока тикет в статусе `pending`
- При получении ответа от пользователя система ищет подходящий `condition`
- Если не нашла - использует default-шаг (если есть)
- Когда доходит до листа (нет детей) → тикет переводится в `open`

### 5.3. Общение после открытия тикета
- Клиент и поддержка могут писать сообщения
- Все сообщения пробрасываются в WebSocket комнату `ticket:{id}`

### 5.4. Закрытие тикета
- Клиент может закрыть свой тикет (`PATCH /support/tickets/{id}/status` с `closed`)
- Поддержка может закрыть назначенный тикет
- После закрытия можно поставить оценку (1–5)

### 5.5. Автоматическое закрытие неактивных тикетов
- В `pending` статусе
- Таймаут: **5 минут**
- При срабатывании бот отправляет сообщение и закрывает тикет

## 6. WebSocket (реал-тайм)

- Эндпоинт: `GET /ws/tickets/{ticket_id}`
- Комнаты: `ticket:{uuid}`
- События:
    - `message_created` (с поддержкой кнопок)
    - `status_changed`
    - `assigned_changed`

## 7. API Эндпоинты

### 7.1. Инициализация контакта (Widget)
- `POST /init/web`
- `POST /init/telegram`

### 7.2. Категории
- `GET /categories` - список (разный для user/support/admin)
- `POST /categories` - только admin
- `PATCH /categories/{id}` - только admin

### 7.3. Тикеты - Клиент
- `POST /tickets`
- `GET /tickets`
- `GET /tickets/{id}`
- `POST /tickets/{id}/messages`
- `GET /tickets/{id}/messages`
- `POST /tickets/{id}/rate`

### 7.4. Тикеты - Поддержка / Админ
- `GET /support/tickets`
- `GET /support/tickets/{id}`
- `PATCH /support/tickets/{id}/assign`
- `PATCH /support/tickets/{id}/status`
- `POST /support/tickets/{id}/messages`
- `GET /support/tickets/{id}/messages`

### 7.5. Сценарии (только admin)
- `POST /scenarios`
- `GET /scenarios`
- `GET /scenarios/{id}`
- `PATCH /scenarios/{id}`
- `DELETE /scenarios/{id}`
- `POST /scenarios/{id}/steps`
- `PATCH /scenarios/{id}/steps/{stepID}`
- `DELETE /scenarios/{id}/steps/{stepID}`

### 7.6. Activity Log (только admin)
- `GET /activity`
- `GET /activity/{ticket_id}`

### 7.7. Swagger UI
- `GET /swagger/index.html`

## 8. Важные нюансы и ограничения

1. **Бот работает только в `pending`** статусе.
2. После перехода в `open` сценарий **не продолжается**.
3. Оценку можно поставить **только** закрытому тикету и **только один раз**.
4. Сообщения нельзя отправлять в `closed` тикет.
5. Поддержка может писать только в назначенные себе тикеты (кроме открытых).
6. У одного сценария может быть **только один** root-шаг.
7. У одного родительского шага может быть **только один** default-переход (без `condition`).
8. Scheduler каждую минуту проверяет неактивные `pending` тикеты.

## 9. Рекомендации по интеграции

- Для Web-виджета сначала вызывайте `POST /init/web`
- Для Telegram - `POST /init/telegram`
- WebSocket подключение нужно делать **после** создания тикета
- При получении `buttons` в сообщении - отображайте их пользователю

---

### 10. Запуск и развертывание проекта

#### 10.1. Настройка окружения (.env)

Создайте файл `.env` в корне проекта на основе `.env.example`:

```env
# Server
SERVER_PORT=8080

# PostgreSQL
POSTGRES_HOST=db
POSTGRES_PORT=5432
POSTGRES_DB=j_support
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your_strong_password

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-in-production

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://your-frontend.com

# WebSocket
WS_ORIGINS=http://localhost:3000,https://your-frontend.com
```

#### 10.2. Запуск через Docker Compose

**Шаг 1:** Убедитесь, что у вас установлен Docker и Docker Compose.

**Шаг 2:** Запустите систему одной командой:

```bash
docker compose up -d --build
```

Приложение будет доступно по адресу:  
**`http://localhost:8080`**

Swagger UI:  
**`http://localhost:8080/swagger/index.html`**

#### 10.3. Миграции базы данных

На текущий момент миграции применяются **вручную**:

```bash
# Применить миграцию вверх
psql -h localhost -p 5438 -U postgres -d j_support -f migrations/000001_init.up.sql

# Откатить миграцию (если нужно)
psql -h localhost -p 5438 -U postgres -d j_support -f migrations/000001_init.down.sql
```

#### 10.4. Переменные окружения (полный список)

| Переменная                | Описание                              | Обязательна | Пример                     |
|---------------------------|---------------------------------------|-------------|----------------------------|
| `SERVER_PORT`             | Порт приложения                       | Да          | 8080                       |
| `POSTGRES_HOST`           | Хост PostgreSQL                       | Да          | db                         |
| `POSTGRES_PORT`           | Порт PostgreSQL                       | Да          | 5432                       |
| `POSTGRES_DB`             | Имя базы данных                       | Да          | j_support                  |
| `POSTGRES_USER`           | Пользователь БД                       | Да          | postgres                   |
| `POSTGRES_PASSWORD`       | Пароль пользователя                   | Да          | strong_password            |
| `JWT_SECRET`              | Секрет для JWT                        | Да          | supersecretkey123          |
| `CORS_ALLOWED_ORIGINS`    | Разрешённые origins через запятую     | Да          | http://localhost:3000      |
| `WS_ORIGINS`              | Разрешённые origins для WebSocket     | Да          | http://localhost:3000      |

#### 10.5. Использование фронтенда

Вы можете запустить веб-сервер на python в директории frontend для обработки запросов.

```bash
cd frontend
```
```bash
python -m http.server 5500
```

После этого веб-интерфейс будет доступен по адресу `http://localhost:5500`

В этой же директории есть 3 файла:
1. `admin.html` - админ-панель для управления тикетами, сценариями и категориями. Требуется вставить jwt-токен в поле.
2. `jwt.html` - пример мобильного приложения. Требует jwt-токен для входа
3. `widget.html` - веб-виджет. После ввода имени и телефона выдаётся уникальный токен.

`⚠️ Веб-интерфейс написан с использованием ИИ-инструментов и некоторые вещи не тестировались / могут не работать`