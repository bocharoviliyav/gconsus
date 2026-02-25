# C4 Architecture Diagrams — GConsus

## L1: System Context

Верхнеуровневый контекст: GConsus как единая система и её внешние взаимодействия.

```mermaid
C4Context
    title GConsus — System Context (L1)

    Person(manager, "Engineering Manager", "Просматривает дашборды, лидерборды, аналитику команд")
    Person(lead, "Tech Lead", "Анализирует метрики команды и разработчиков, управляет составом")
    Person(admin, "Admin", "Настраивает VCS-провайдеры, управляет пользователями и синхронизацией")

    System(gconsus, "GConsus", "Аналитическая платформа визуализации активности разработчиков")

    System_Ext(github, "GitHub Enterprise", "Хостинг репозиториев, PR, issues, reviews")
    System_Ext(gitlab, "GitLab", "Хостинг репозиториев, MR, issues, reviews")
    System_Ext(hr_api, "Employee HR API", "Корпоративный справочник сотрудников")
    System_Ext(ml_service, "ML Service", "Предиктивная аналитика (будущее)")

    Rel(manager, gconsus, "Просматривает аналитику", "HTTPS")
    Rel(lead, gconsus, "Управляет командой, смотрит метрики", "HTTPS")
    Rel(admin, gconsus, "Настраивает систему", "HTTPS")

    Rel(gconsus, github, "Получает коммиты, PR, reviews, issues", "GraphQL + REST")
    Rel(gconsus, gitlab, "Получает коммиты, MR, reviews, issues", "REST API v4")
    Rel(gconsus, hr_api, "Синхронизирует сотрудников", "REST")
    Rel(gconsus, ml_service, "Запрашивает прогнозы", "REST")
```

## L2: Container Diagram

Внутренние контейнеры системы и их связи.

```mermaid
C4Container
    title GConsus — Container Diagram (L2)

    Person(user, "Пользователь", "Manager / Lead / Admin")

    System_Boundary(gconsus, "GConsus Platform") {
        Container(frontend, "Frontend", "React 18, TypeScript, Vite, Tailwind CSS", "SPA с дашбордами, лидербордами, графиками. Nginx в Docker")
        Container(backend, "Backend API", "Go 1.24, net/http", "REST API сервер. Сбор данных из VCS, агрегация метрик, RBAC")
        ContainerDb(postgres, "PostgreSQL 16", "SQL", "users, teams, git_activities, aggregated_metrics, configurations, sync_history")
        Container(keycloak, "Keycloak 23", "Java, OIDC", "SSO, аутентификация, управление ролями (admin, manager, user)")
        Container(wiremock, "WireMock VCS", "WireMock", "Мок-сервер GitHub/GitLab API для разработки")
    }

    System_Ext(github, "GitHub Enterprise", "REST + GraphQL API")
    System_Ext(gitlab, "GitLab", "REST API v4")
    System_Ext(hr_api, "Employee HR API", "REST")
    System_Ext(ml_service, "ML Service", "FastAPI (будущее)")

    Rel(user, frontend, "Использует", "HTTPS, порт 80")
    Rel(frontend, backend, "Запросы API", "HTTP/JSON, /api/v1/*")
    Rel(frontend, keycloak, "Аутентификация", "OIDC, JS Adapter")

    Rel(backend, postgres, "Читает/пишет данные", "TCP, sqlx/lib/pq")
    Rel(backend, keycloak, "Валидация JWT", "HTTP")
    Rel(backend, github, "Fetch activities", "GraphQL + REST")
    Rel(backend, gitlab, "Fetch activities", "REST API v4")
    Rel(backend, hr_api, "Sync employees", "REST")
    Rel(backend, ml_service, "Predict", "REST")
    Rel(backend, wiremock, "Dev: имитация VCS", "HTTP")

    UpdateLayoutConfig($c4ShapeInRow="3", $c4BoundaryInRow="1")
```

## L3: Component Diagram (Backend)

Внутренняя структура бэкенд-контейнера: слои handler → service → repository → adapter.

```mermaid
C4Component
    title GConsus Backend — Component Diagram (L3)

    Container_Boundary(backend, "Backend (Go 1.24)") {

        Component(server, "HTTP Server", "net/http, ServeMux", "Роутинг, CORS, middleware (JWT валидация, RBAC)")

        Component(users_h, "UsersHandler", "handler/users.go", "CRUD пользователей: POST/GET/PUT/DELETE /users")
        Component(teams_h, "TeamsHandler", "handler/teams.go", "CRUD команд + управление участниками")
        Component(analytics_h, "AnalyticsHandler", "handler/analytics.go", "Dashboard, leaderboard, user/team/repo analytics")
        Component(settings_h, "SettingsHandler", "handler/settings.go", "VCS-провайдеры, системные настройки")
        Component(sync_h, "SyncHandler", "handler/sync.go", "Статус и ручной запуск синхронизации")

        Component(user_svc, "UserService", "service/users.go", "Бизнес-логика пользователей")
        Component(team_svc, "TeamService", "service/team_service.go", "Бизнес-логика команд и участников")
        Component(analytics_svc, "AnalyticsService", "service/analytics_service.go", "Расчёт аналитики, лидерборды, скоринг")
        Component(agg_svc, "AggregationService", "service/aggregation.go", "Агрегация метрик по пользователям и командам")
        Component(settings_svc, "SettingsService", "service/settings_service.go", "Управление провайдерами и конфигурацией")
        Component(sync_svc, "SyncService", "service/sync_service.go", "Синхронизация данных из VCS-провайдеров")
        Component(scheduler, "Scheduler", "service/scheduler.go", "Cron-задачи: sync, aggregation, employee sync")

        Component(user_repo, "UserRepository", "repository/user_repository.go", "SQL: users")
        Component(team_repo, "TeamRepository", "repository/team_repository.go", "SQL: teams, team_members")
        Component(activity_repo, "ActivityRepository", "repository/activity_repository.go", "SQL: git_activities")
        Component(metrics_repo, "MetricsRepository", "repository/metrics_repository.go", "SQL: aggregated_metrics")
        Component(provider_repo, "ProviderRepository", "repository/provider_repository.go", "SQL: vcs_providers")
        Component(sync_repo, "SyncRepository", "repository/sync_repository.go", "SQL: sync_history, configurations")

        Component(github_adapter, "GitHub Adapter", "adapter/github/", "GraphQL contributionsCollection + REST API")
        Component(gitlab_adapter, "GitLab Adapter", "adapter/gitlab/", "REST API v4 (commits, MR, issues, reviews)")
        Component(vcs_iface, "vcs.Client Interface", "adapter/vcs/", "Унифицированный интерфейс VCS-адаптера")
        Component(ml_connector, "ML Connector", "adapter/ml/", "Stub-коннектор к ML-сервису (неактивен)")
    }

    ContainerDb(db, "PostgreSQL 16", "SQL")
    Container_Ext(kc, "Keycloak", "OIDC")
    System_Ext(gh, "GitHub", "API")
    System_Ext(gl, "GitLab", "API")
    System_Ext(ml, "ML Service", "API")

    Rel(server, users_h, "Роутинг")
    Rel(server, teams_h, "Роутинг")
    Rel(server, analytics_h, "Роутинг")
    Rel(server, settings_h, "Роутинг")
    Rel(server, sync_h, "Роутинг")

    Rel(users_h, user_svc, "Вызывает")
    Rel(teams_h, team_svc, "Вызывает")
    Rel(analytics_h, analytics_svc, "Вызывает")
    Rel(settings_h, settings_svc, "Вызывает")
    Rel(sync_h, sync_svc, "Вызывает")

    Rel(analytics_svc, agg_svc, "CalculateScore")
    Rel(scheduler, sync_svc, "Cron: sync activities")
    Rel(scheduler, agg_svc, "Cron: aggregate metrics")

    Rel(user_svc, user_repo, "SQL")
    Rel(team_svc, team_repo, "SQL")
    Rel(analytics_svc, metrics_repo, "SQL")
    Rel(analytics_svc, activity_repo, "SQL")
    Rel(analytics_svc, user_repo, "SQL")
    Rel(analytics_svc, team_repo, "SQL")
    Rel(agg_svc, activity_repo, "SQL")
    Rel(agg_svc, metrics_repo, "SQL")
    Rel(settings_svc, provider_repo, "SQL")
    Rel(sync_svc, activity_repo, "SQL")
    Rel(sync_svc, sync_repo, "SQL")

    Rel(sync_svc, vcs_iface, "FetchUserActivities")
    Rel(vcs_iface, github_adapter, "implements")
    Rel(vcs_iface, gitlab_adapter, "implements")
    Rel(analytics_svc, ml_connector, "Predict (future)")

    Rel(user_repo, db, "sqlx")
    Rel(team_repo, db, "sqlx")
    Rel(activity_repo, db, "sqlx")
    Rel(metrics_repo, db, "sqlx")
    Rel(provider_repo, db, "sqlx")
    Rel(sync_repo, db, "sqlx")

    Rel(github_adapter, gh, "GraphQL + REST")
    Rel(gitlab_adapter, gl, "REST v4")
    Rel(ml_connector, ml, "REST")
    Rel(server, kc, "JWT validation")
```
