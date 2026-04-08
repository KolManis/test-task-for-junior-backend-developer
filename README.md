# Task Service

Сервис для управления задачами с HTTP API на Go.

## Требования

- Go `1.23+`
- Docker и Docker Compose

## Быстрый запуск через Docker Compose

```bash
docker compose up --build
```

После запуска сервис будет доступен по адресу `http://localhost:8080`.

Если `postgres` уже запускался ранее со старой схемой, пересоздай volume:

```bash
docker compose down -v
docker compose up --build
```

Причина в том, что SQL-файл из `migrations/0001_create_tasks.up.sql` монтируется в `docker-entrypoint-initdb.d` и применяется только при инициализации пустого data volume.

## Swagger

Swagger UI:

```text
http://localhost:8080/swagger/
```

OpenAPI JSON:

```text
http://localhost:8080/swagger/openapi.json
```

## API

Базовый префикс API:

```text
/api/v1
```

Основные маршруты:

- `POST /api/v1/tasks`
- `GET /api/v1/tasks`
- `GET /api/v1/tasks/{id}`
- `PUT /api/v1/tasks/{id}`
- `DELETE /api/v1/tasks/{id}`

## Принятые решения и допущения

### Архитектура

Периодичность вынесена в отдельную сущность `recurrence_rules`, а не добавлена как поле в `tasks`. Это позволяет:
- Хранить историю изменений правил
- Легко расширять новыми типами периодичности
- Удалять правило без потери уже выполненных задач (связь `ON DELETE SET NULL`)

### Генерация задач

Задачи генерируются **моментально при создании правила**, а не вычисляются на лету при запросе. Это обеспечивает:
- Возможность редактировать отдельные экземпляры периодических задач
- Простой и быстрый `GET /tasks` без сложной логики
- Прозрачную связь "правило → задачи" через `recurrence_id`

### Ограничения

- **Максимум 500 задач** на одно правило — защита от случайного создания тысяч задач
- **Диапазон дат обязателен** — пользователь должен явно указать горизонт планирования
- **`day_of_month` ограничен 1–30** — согласно требованию ТЗ, 31-е число пропускается
- **Даты вне диапазона в `specific_dates` игнорируются** — удобнее, чем ошибка 400

### Обработка граничных случаев

| Случай | Решение |
|--------|---------|
| 30 февраля | Go автоматически нормализует дату → проверяем `current.Day() == day`, задача не создаётся |
| Дубликаты в `specific_dates` | Отфильтровываются через `map[string]bool` |
| Пустой список `specific_dates` | Ошибка валидации |
| `end_date < start_date` | Ошибка на уровне БД и в коде |
| Непереданный `every_n_days` | По умолчанию = 1 (каждый день) |


## API ручки для периодичности

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/api/v1/recurrence-rules` | Создать правило периодичности и сгенерировать задачи |
| `GET` | `/api/v1/recurrence-rules` | Список всех правил периодичности |
| `GET` | `/api/v1/recurrence-rules/{id}` | Получить правило по ID |
| `DELETE` | `/api/v1/recurrence-rules/{id}` | Удалить правило и все связанные с ним задачи |

---

## Формат запросов и ответов

### `POST /api/v1/recurrence-rules`

**Тело запроса:**
```json
{
    "task": {
        "title": "Название задачи",
        "description": "Описание (опционально)",
        "status": "new|in_progress|done"
    },
    "type": "daily|monthly_day|specific_dates|even_odd_days",
    "params": {
        // зависит от type
    },
    "start_date": "2026-04-01",
    "end_date": "2026-04-30"
}
```

**Ответ (201 Created):**
```json
{
    "rule": {
        "id": 1,
        "type": "daily",
        "params": {...},
        "start_date": "2026-04-01",
        "end_date": "2026-04-30",
        "created_at": "2026-04-08T11:52:03Z"
    },
    "tasks": [
        {
            "id": 1,
            "title": "...",
            "description": "...",
            "status": "new",
            "recurrence_id": 1,
            "scheduled_date": "2026-04-01",
            "created_at": "...",
            "updated_at": "..."
        }
    ]
}
```

---

### `GET /api/v1/recurrence-rules`

**Ответ (200 OK):**
```json
[
    {
        "id": 1,
        "type": "daily",
        "params": {"every_n_days": 2},
        "start_date": "2026-04-01",
        "end_date": "2026-04-30",
        "created_at": "2026-04-08T11:52:03Z"
    }
]
```

---

### `GET /api/v1/recurrence-rules/{id}`

**Ответ (200 OK):**
```json
{
    "id": 1,
    "type": "daily",
    "params": {"every_n_days": 2},
    "start_date": "2026-04-01",
    "end_date": "2026-04-30",
    "created_at": "2026-04-08T11:52:03Z"
}
```

**Ошибка (404):**
```json
{"error": "recurrence rule not found"}
```

---

### `DELETE /api/v1/recurrence-rules/{id}`

**Ответ:** `204 No Content` (пустое тело)

---

## Параметры для разных типов

| `type` | `params` | Пример |
|--------|----------|--------|
| `daily` | `{"every_n_days": N}` | `{"every_n_days": 2}` |
| `monthly_day` | `{"day_of_month": N}` | `{"day_of_month": 15}` |
| `specific_dates` | `{"dates": ["YYYY-MM-DD"]}` | `{"dates": ["2026-04-10", "2026-04-20"]}` |
| `even_odd_days` | `{"parity": "even"}` или `"odd"` | `{"parity": "even"}` |

---
