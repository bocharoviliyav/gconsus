# GConsus -- техническая документация

## 1. Назначение системы

GConsus -- аналитическая платформа для сбора, хранения и визуализации данных об активности разработчиков в системах контроля версий (GitHub Enterprise, GitLab). Система предоставляет дашборды, лидерборды и детальную аналитику на уровне пользователей, команд и репозиториев.

## 2. Архитектура

### 2.1 Компоненты

- **Backend** (Go 1.24) -- REST API сервер, сбор данных из VCS, агрегация метрик.
- **Frontend** (React + TypeScript + Vite) -- веб-интерфейс с графиками и дашбордами.
- **PostgreSQL 16** -- хранилище пользователей, активностей, агрегированных метрик и конфигурации.
- **Keycloak 23** -- аутентификация и управление доступом (OIDC).
- **WireMock VCS** -- мок-сервер, имитирующий реальные API GitHub и GitLab для разработки.

### 2.2 Структура каталогов

```
backend/
  adapter/
    github/     -- GitHub REST + GraphQL клиент (vcs.Client)
    gitlab/     -- GitLab API v4 клиент (vcs.Client)
    vcs/        -- унифицированный интерфейс VCS адаптера
    ml/         -- неактивный коннектор к ML-сервису предиктивной аналитики
  entity/       -- доменные модели и DTO
  handler/      -- HTTP хендлеры (users, teams, analytics, settings, sync)
  lib/          -- утилиты (middleware, logging, rest helpers)
  repository/   -- слой доступа к PostgreSQL (activity, team, user, metrics, provider, sync)
  service/      -- бизнес-логика (analytics, aggregation, settings, sync, user, team)
  database/     -- подключение к БД (sqlx)
  main.go       -- точка входа и DI
frontend/
  src/
    services/api/  -- API клиенты (analytics, dashboard, teams, settings, repositories)
    types/         -- TypeScript типы
    components/    -- React компоненты
migrations/
  000001_init_schema.up.sql  -- начальная схема БД
wiremock/                    -- legacy моки (имитация бэкенд API)
wiremock-vcs/                -- моки реальных VCS API (GitHub REST/GraphQL, GitLab v4)
docs/                        -- документация
```

## 3. Схема сбора и обработки данных

### 3.1 Поток данных

1. **Регистрация провайдеров.** Администратор через API `/settings/providers` добавляет VCS-провайдеры (тип, URL, токен).
2. **Синхронизация.** SyncService итерирует по всем активным провайдерам и пользователям. Для каждой пары (провайдер, пользователь) вызывается `vcs.Client.FetchUserActivities`.
3. **Сохранение.** Полученные активности (коммиты, PR, ревью, задачи) преобразуются в `entity.GitActivity` и пакетно вставляются в таблицу `git_activities`.
4. **Агрегация.** AggregationService рассчитывает сводные метрики по пользователям и командам, записывает их в `aggregated_metrics`.
5. **Визуализация.** Frontend запрашивает данные через REST API, строит графики и лидерборды.

### 3.2 Интерфейс VCS адаптера

Файл: `backend/adapter/vcs/interface.go`

Любой VCS провайдер реализует интерфейс `vcs.Client`:

- `FetchUserActivities(ctx, username, from, to)` -- все активности пользователя за период.
- `FetchRepositories(ctx, org)` -- список репозиториев.
- `FetchPullRequestStats(ctx, owner, repo, prNumber)` -- детали PR.
- `TestConnection(ctx)` -- проверка подключения.

Пагинация обрабатывается внутри адаптера. Вызывающий код получает полный набор данных.

### 3.3 GitHub адаптер

- Коммиты, PR, задачи, ревью -- через GraphQL `contributionsCollection`.
- PR обогащаются данными (additions, deletions) через REST `/repos/{owner}/{repo}/pulls/{number}`.
- Пагинация REST -- по заголовку `Link: <...>; rel="next"`.

### 3.4 GitLab адаптер

- Коммиты -- через REST `/projects/{id}/repository/commits` с фильтром `since/until`.
- Merge Requests -- `/merge_requests` с фильтром `author_id`.
- Ревью -- `/merge_requests` с фильтром `reviewer_id`.
- Задачи -- `/issues` с фильтром `author_id`.
- Пагинация -- по заголовку `X-Next-Page`.

## 4. Схема базы данных

### 4.1 Основные таблицы

- `users` -- сотрудники (username, ФИО, email, должность).
- `teams` -- команды разработки.
- `team_members` -- связь пользователь-команда с ролью и датами.
- `vcs_providers` -- зарегистрированные VCS провайдеры (type: github/gitlab).
- `git_activities` -- единичные активности (commit, pr, review, issue) с привязкой к пользователю и провайдеру.
- `aggregated_metrics` -- предрасчитанные метрики за период (по пользователю или по команде).
- `configurations` -- системные настройки (расписание синхронизации, retention и пр.).
- `sync_history` -- журнал запусков синхронизации.

## 5. API Reference

Базовый путь: `/api/v1`

### 5.1 Пользователи

| Метод | Путь | Описание |
|-------|------|----------|
| POST | /users | Создание пользователя |
| GET | /users | Список пользователей |
| GET | /users/{id} | Получение пользователя |
| PUT | /users/{id} | Обновление пользователя |
| DELETE | /users/{id} | Деактивация пользователя |

### 5.2 Команды

| Метод | Путь | Описание |
|-------|------|----------|
| POST | /teams | Создание команды |
| GET | /teams | Список команд |
| GET | /teams/{id} | Детали команды с участниками |
| PUT | /teams/{id} | Обновление команды |
| DELETE | /teams/{id} | Деактивация команды |
| POST | /teams/{id}/members | Добавление участника |
| DELETE | /teams/{teamId}/members/{userId} | Удаление участника |

### 5.3 Аналитика

| Метод | Путь | Описание |
|-------|------|----------|
| GET | /analytics/dashboard?start=&end= | Сводная статистика |
| GET | /analytics/leaderboard?start=&end=&limit= | Лидерборд пользователей |
| GET | /analytics/teams/leaderboard?start=&end= | Лидерборд команд |
| GET | /analytics/users/{id}?start=&end= | Аналитика пользователя |
| GET | /analytics/teams/{id}?start=&end= | Аналитика команды |
| GET | /analytics/repositories?owner=&name= | Аналитика репозитория |
| GET | /analytics/repositories/leaderboard?start=&end= | Лидерборд репозиториев |

Параметры дат: `start`, `end` в формате RFC3339. По умолчанию -- последние 30 дней.

### 5.4 Настройки

| Метод | Путь | Описание |
|-------|------|----------|
| GET | /settings/system | Системные настройки и статус |
| PUT | /settings/system | Обновление настроек |
| GET | /settings/providers | Список VCS провайдеров |
| POST | /settings/providers | Добавление провайдера |
| POST | /settings/test-github | Проверка подключения к GitHub |
| POST | /settings/test-gitlab | Проверка подключения к GitLab |

### 5.5 Синхронизация

| Метод | Путь | Описание |
|-------|------|----------|
| GET | /sync/status | Статус последней синхронизации |
| POST | /sync/trigger | Запуск ручной синхронизации |
| GET | /sync/history?page=&page_size= | История синхронизаций |

### 5.6 Legacy

| Метод | Путь | Описание |
|-------|------|----------|
| GET | /users/{login}/activity?from=&to= | Активность пользователя (GraphQL) |

## 6. Развертывание

### 6.1 Требования

- Docker и Docker Compose v2.
- Порты: 80 (frontend), 8000 (backend), 5432 (PostgreSQL), 8090 (Keycloak), 9090 (WireMock VCS).

### 6.2 Запуск в режиме разработки

```bash
# Полный стек (backend берет данные из WireMock VCS)
docker compose up -d

# Только legacy-режим (фронт работает через WireMock без бэкенда)
docker compose --profile legacy-mocks up -d wiremock frontend
```

### 6.3 Переменные окружения

Файл `.env` содержит все настройки. Ключевые переменные:

- `DB_PASSWORD` -- пароль PostgreSQL.
- `ML_ENABLED`, `ML_ENDPOINT` -- коннектор к ML сервису (по умолчанию выключен).
- `VITE_API_URL` -- адрес бэкенда для фронтенда.

### 6.4 Миграции

Миграции выполняются автоматически при первом запуске PostgreSQL. SQL-файлы расположены в `migrations/` и монтируются в контейнер через `docker-entrypoint-initdb.d`.

### 6.5 Production

Для production-окружения:

1. Настроить VCS провайдеры (GitHub/GitLab) через Settings UI или API (`/settings/providers`).
2. Установить `ML_ENABLED=true` при наличии ML-сервиса.
3. Убрать `wiremock-vcs` из `depends_on` бэкенда.
4. Настроить CORS (`CORS_ALLOWED_ORIGINS`).
5. Настроить Keycloak realm и клиенты.

## 7. ML коннектор

Неактивный коннектор к внешнему ML-сервису (`backend/adapter/ml/connector.go`). Предназначен для будущей интеграции с моделями предиктивной аналитики. Подробнее -- в документе `docs/predictive-analytics.md`.

Интерфейс `ml.Client`:

- `Predict(ctx, PredictionRequest)` -- запрос прогноза.
- `HealthCheck(ctx)` -- проверка доступности.
- `IsEnabled()` -- статус коннектора.

Активация: `ML_ENABLED=true`, `ML_ENDPOINT=http://ml-service:8501`.
