# 📊 GConsus

Комплексная платформа для аналитики и оценки эффективности разработчиков с поддержкой GitHub и GitLab.

## 🏗️ Архитектура системы

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Frontend      │    │    Keycloak      │    │   PostgreSQL    │
│   (React)       │    │  (Authentication)│    │   (Database)    │
│   Port: 80      │    │   Port: 8090     │    │   Port: 5432    │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                        │                       │
         └────────────────────────┼───────────────────────┘
                                  │
                     ┌─────────────────┐    ┌─────────────────┐
                     │    Backend      │    │    WireMock     │
                     │     (Go)        │    │   (Mock API)    │
                     │   Port: 8080    │    │   Port: 8080    │
                     │(закомментирован)│    └─────────────────┘
                     └─────────────────┘
```

## 🚀 Быстрый старт

```bash
# Клонирование и запуск
git clone <repository-url>

# Запуск всех сервисов
docker compose up -d

# Проверка статуса
docker compose ps
```

## 🌐 Доступные сервисы

| Сервис | URL | Статус | Описание |
|--------|-----|--------|----------|
| **Frontend** | http://localhost:80 | ✅ | React приложение |
| **Keycloak Admin** | http://localhost:8090/admin | ✅ | Админ-панель (admin/admin123) |
| **Database** | localhost:5432 | ✅ | PostgreSQL |
| **WireMock** | http://localhost:8080 | ✅ | Mock API |

## 🔐 Тестовые пользователи

| Логин | Пароль | Роли |
|-------|--------|------|
| admin | admin123 | admin, manager, user |
| manager | manager123 | manager, user |
| testuser | user123 | user |

## 📋 Описание

GConsus- это корпоративное решение для руководителей всех уровней, предоставляющее детальную аналитику и метрики эффективности разработчиков. Система автоматически собирает данные из VCS систем (GitHub, GitLab), агрегирует их и отображает в виде интерактивных дашбордов и лидербордов.

## ⚙️ Конфигурация

### Переменные окружения

```env
# Database
DB_PASSWORD=changeme123

# Keycloak
KEYCLOAK_ADMIN_PASSWORD=admin123
KEYCLOAK_REALM=gconsus
KEYCLOAK_CLIENT_ID=gconsus-backend

# Frontend
VITE_KEYCLOAK_URL=http://localhost:8090
VITE_KEYCLOAK_REALM=gconsus
VITE_KEYCLOAK_CLIENT_ID=gconsus-frontend
```



### Ключевые возможности

- 📊 **Лидерборды**: Рейтинги разработчиков и команд по различным метрикам
- 👥 **Управление командами**: Создание команд, добавление участников, назначение ролей
- 📈 **Детальная аналитика**: Метрики по коммитам, PRs, reviews, issues
- 🔄 **Автоматическая синхронизация**: Периодическая загрузка данных из GitHub/GitLab
- 🔐 **Аутентификация через Keycloak**: SSO с RBAC (admin, manager, user)
- 🌍 **Мультиязычность**: Русский и английский интерфейс
- 🌓 **Темная/светлая тема**: Адаптивный дизайн



### Технологический стек

**Backend:**
- Go 1.24
- PostgreSQL 16
- sqlx для работы с БД
- Keycloak для аутентификации
- robfig/cron для фоновых задач

**Frontend:**
- React 18 + TypeScript
- Bun (package manager & bundler)
- Vite (build tool)
- Tailwind CSS
- React-i18next

**Infrastructure:**
- Docker & Docker Compose
- Nginx (для frontend)
- Makefile для CI/CD

## 🛠️ Разработка

### Основные команды

```bash
make help              # Показать все команды
make build             # Собрать backend и frontend
make test              # Запустить тесты
make lint              # Проверить код
make build-images      # Собрать Docker images
make db-shell          # Открыть PostgreSQL shell
make logs              # Показать логи
```
