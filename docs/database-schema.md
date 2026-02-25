# Database Schema Design

## ER Diagram

```mermaid
erDiagram
    users {
        uuid id PK
        varchar username UK "unique, max 100"
        varchar first_name
        varchar last_name
        varchar patronymic "nullable"
        varchar email "nullable"
        text photo_url "nullable"
        varchar position "nullable"
        boolean is_active "default true"
        timestamptz created_at
        timestamptz updated_at
    }

    teams {
        uuid id PK
        varchar name UK "unique"
        text description "nullable"
        uuid manager_id FK "nullable, -> users"
        boolean is_active "default true"
        timestamptz created_at
        timestamptz updated_at
    }

    team_members {
        uuid id PK
        uuid team_id FK "-> teams, ON DELETE CASCADE"
        uuid user_id FK "-> users, ON DELETE CASCADE"
        varchar role "default developer"
        timestamptz joined_at
        timestamptz left_at "nullable"
    }

    vcs_providers {
        uuid id PK
        varchar name
        varchar type "github | gitlab"
        varchar base_url
        text auth_token
        boolean enabled "default true"
        timestamptz created_at
        timestamptz updated_at
    }

    git_activities {
        uuid id PK
        uuid user_id FK "-> users, ON DELETE CASCADE"
        uuid provider_id FK "-> vcs_providers, ON DELETE CASCADE"
        varchar activity_type "commit | pr | issue | review"
        varchar repository_name
        varchar repository_owner
        integer commit_count "default 0"
        integer lines_added "default 0"
        integer lines_deleted "default 0"
        text pr_title "nullable"
        text pr_url "nullable"
        boolean pr_merged "nullable"
        text issue_title "nullable"
        text issue_url "nullable"
        varchar issue_state "nullable"
        timestamptz occurred_at
        timestamptz fetched_at
        jsonb raw_data "nullable, GIN index"
    }

    aggregated_metrics {
        uuid id PK
        uuid user_id FK "nullable, -> users"
        uuid team_id FK "nullable, -> teams"
        date period_start
        date period_end
        integer total_commits "default 0"
        bigint total_lines_added "default 0"
        bigint total_lines_deleted "default 0"
        integer total_prs "default 0"
        integer total_prs_merged "default 0"
        integer total_reviews "default 0"
        integer total_issues "default 0"
        integer repositories_count "default 0"
        jsonb top_repositories "nullable"
        timestamptz created_at
    }

    configurations {
        uuid id PK
        varchar key UK "unique"
        jsonb value
        text description "nullable"
        timestamptz updated_at
        uuid updated_by FK "nullable, -> users"
    }

    sync_history {
        uuid id PK
        uuid provider_id FK "nullable, -> vcs_providers"
        varchar sync_type "employees | git_activities | aggregation"
        varchar status "running | completed | failed"
        integer users_synced "default 0"
        integer activities_synced "default 0"
        timestamptz started_at
        timestamptz completed_at "nullable"
        text error_message "nullable"
    }

    users ||--o{ team_members : "member of"
    teams ||--o{ team_members : "has members"
    users ||--o| teams : "manages (manager_id)"
    users ||--o{ git_activities : "performs"
    vcs_providers ||--o{ git_activities : "source"
    users ||--o{ aggregated_metrics : "metrics per user"
    teams ||--o{ aggregated_metrics : "metrics per team"
    vcs_providers ||--o{ sync_history : "sync logs"
    users ||--o{ configurations : "updated_by"
```

## Table Descriptions

### users
Stores employee information from external HR API and Git systems.

**Columns:**
- `id`: UUID primary key
- `username`: VCS username (unique, up to 100 chars)
- `first_name`: Employee first name
- `last_name`: Employee last name
- `patronymic`: Employee patronymic (optional)
- `email`: Work email
- `photo_url`: URL to employee photo
- `position`: Job title
- `is_active`: Employment status
- `created_at`, `updated_at`: Timestamps

**Indexes:**
- `idx_users_username` (unique)
- `idx_users_email`
- `idx_users_is_active`

---

### teams
Developer teams/squads organization.

**Columns:**
- `id`: UUID primary key
- `name`: Team name (unique)
- `description`: Team description
- `manager_id`: Foreign key to users (team lead)
- `is_active`: Team status
- `created_at`, `updated_at`: Timestamps

**Indexes:**
- `idx_teams_name` (unique)
- `idx_teams_manager_id`

---

### team_members
Many-to-many relationship between users and teams.

**Columns:**
- `id`: UUID primary key
- `team_id`: Foreign key to teams
- `user_id`: Foreign key to users
- `role`: Member role (developer, lead, etc.)
- `joined_at`: When user joined the team
- `left_at`: When user left (NULL if active)

**Indexes:**
- `idx_team_members_team_id`
- `idx_team_members_user_id`
- `idx_team_members_active` (where left_at IS NULL)
- `uniq_team_user` (unique on team_id, user_id, left_at)

---

### vcs_providers
VCS system configurations (GitHub, GitLab).

**Columns:**
- `id`: UUID primary key
- `name`: Provider name (e.g., "GitHub Enterprise", "GitLab Main")
- `type`: Provider type (github, gitlab)
- `base_url`: API base URL
- `auth_token`: Encrypted authentication token
- `enabled`: Whether provider is active
- `created_at`, `updated_at`: Timestamps

**Indexes:**
- `idx_vcs_providers_type`
- `idx_vcs_providers_enabled`

---

### git_activities
Raw activity data from VCS providers.

**Columns:**
- `id`: UUID primary key
- `user_id`: Foreign key to users
- `provider_id`: Foreign key to vcs_providers
- `activity_type`: Type (commit, pr, issue, review)
- `repository_name`: Repository name
- `repository_owner`: Repository owner
- `commit_count`: Number of commits (for commit type)
- `lines_added`: Lines of code added
- `lines_deleted`: Lines of code deleted
- `pr_title`: Pull request title (for pr type)
- `pr_url`: Pull request URL
- `pr_merged`: Whether PR was merged
- `issue_title`: Issue title (for issue type)
- `issue_url`: Issue URL
- `issue_state`: Issue state
- `occurred_at`: When activity happened
- `fetched_at`: When we fetched this data
- `raw_data`: JSONB with full API response

**Indexes:**
- `idx_git_activities_user_id`
- `idx_git_activities_provider_id`
- `idx_git_activities_type`
- `idx_git_activities_occurred_at`
- `idx_git_activities_repository`
- `idx_git_activities_raw_data` (GIN index for JSONB)

---

### aggregated_metrics
Pre-calculated metrics for performance.

**Columns:**
- `id`: UUID primary key
- `user_id`: Foreign key to users (NULL for team-level)
- `team_id`: Foreign key to teams (NULL for user-level)
- `period_start`: Period start date
- `period_end`: Period end date
- `total_commits`: Total commits count
- `total_lines_added`: Total lines added
- `total_lines_deleted`: Total lines deleted
- `total_prs`: Total pull requests
- `total_prs_merged`: Total merged PRs
- `total_reviews`: Total code reviews
- `total_issues`: Total issues created
- `repositories_count`: Unique repositories count
- `top_repositories`: JSONB with top repositories and their stats
- `created_at`: When metrics were calculated

**Indexes:**
- `idx_aggregated_metrics_user_id`
- `idx_aggregated_metrics_team_id`
- `idx_aggregated_metrics_period`
- `uniq_aggregated_metrics` (unique on user_id, team_id, period_start, period_end)

---

### configurations
Application configuration key-value store.

**Columns:**
- `id`: UUID primary key
- `key`: Configuration key (unique)
- `value`: Configuration value (JSONB)
- `description`: Human-readable description
- `updated_at`: Last update timestamp
- `updated_by`: User ID who updated

**Indexes:**
- `idx_configurations_key` (unique)

**Example configs:**
- `sync_schedule`: Cron expression for sync jobs
- `employee_api_url`: External HR API endpoint
- `default_period_days`: Default analytics period

---

### sync_history
History of synchronization jobs.

**Columns:**
- `id`: UUID primary key
- `provider_id`: Foreign key to vcs_providers (NULL for employee sync)
- `sync_type`: Type (employees, git_activities, aggregation)
- `status`: Status (running, completed, failed)
- `users_synced`: Number of users synced
- `activities_synced`: Number of activities synced
- `started_at`: Job start time
- `completed_at`: Job completion time
- `error_message`: Error details if failed

**Indexes:**
- `idx_sync_history_provider_id`
- `idx_sync_history_type`
- `idx_sync_history_status`
- `idx_sync_history_started_at`

---

## Partitioning Strategy

For large datasets, consider partitioning `git_activities` by `occurred_at` (monthly or quarterly partitions).

## Retention Policy

- `git_activities`: Keep raw data for 2 years
- `aggregated_metrics`: Keep indefinitely
- `sync_history`: Keep for 6 months
