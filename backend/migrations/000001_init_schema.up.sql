-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(100) NOT NULL UNIQUE,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    patronymic VARCHAR(100),
    email VARCHAR(255),
    photo_url TEXT,
    position VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_is_active ON users(is_active);

-- Create teams table
CREATE TABLE teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    manager_id UUID REFERENCES users(id) ON DELETE SET NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_teams_name ON teams(name);
CREATE INDEX idx_teams_manager_id ON teams(manager_id);
CREATE INDEX idx_teams_is_active ON teams(is_active);

-- Create team_members table
CREATE TABLE team_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) DEFAULT 'developer',
    joined_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    left_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_team_members_team_id ON team_members(team_id);
CREATE INDEX idx_team_members_user_id ON team_members(user_id);
CREATE INDEX idx_team_members_active ON team_members(team_id, user_id) WHERE left_at IS NULL;
CREATE UNIQUE INDEX uniq_team_user_active ON team_members(team_id, user_id) WHERE left_at IS NULL;

-- Create vcs_providers table
CREATE TABLE vcs_providers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('github', 'gitlab')),
    base_url VARCHAR(500) NOT NULL,
    auth_token TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_vcs_providers_type ON vcs_providers(type);
CREATE INDEX idx_vcs_providers_enabled ON vcs_providers(enabled);

-- Create git_activities table
CREATE TABLE git_activities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider_id UUID NOT NULL REFERENCES vcs_providers(id) ON DELETE CASCADE,
    activity_type VARCHAR(20) NOT NULL CHECK (activity_type IN ('commit', 'pr', 'issue', 'review')),
    repository_name VARCHAR(255) NOT NULL,
    repository_owner VARCHAR(255) NOT NULL,
    commit_count INTEGER DEFAULT 0,
    lines_added INTEGER DEFAULT 0,
    lines_deleted INTEGER DEFAULT 0,
    pr_title TEXT,
    pr_url TEXT,
    pr_merged BOOLEAN,
    issue_title TEXT,
    issue_url TEXT,
    issue_state VARCHAR(50),
    occurred_at TIMESTAMP WITH TIME ZONE NOT NULL,
    fetched_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    raw_data JSONB
);

CREATE INDEX idx_git_activities_user_id ON git_activities(user_id);
CREATE INDEX idx_git_activities_provider_id ON git_activities(provider_id);
CREATE INDEX idx_git_activities_type ON git_activities(activity_type);
CREATE INDEX idx_git_activities_occurred_at ON git_activities(occurred_at DESC);
CREATE INDEX idx_git_activities_repository ON git_activities(repository_owner, repository_name);
CREATE INDEX idx_git_activities_raw_data ON git_activities USING GIN(raw_data);
CREATE INDEX idx_git_activities_user_occurred ON git_activities(user_id, occurred_at DESC);

-- Create aggregated_metrics table
CREATE TABLE aggregated_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    team_id UUID REFERENCES teams(id) ON DELETE CASCADE,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    total_commits INTEGER NOT NULL DEFAULT 0,
    total_lines_added BIGINT NOT NULL DEFAULT 0,
    total_lines_deleted BIGINT NOT NULL DEFAULT 0,
    total_prs INTEGER NOT NULL DEFAULT 0,
    total_prs_merged INTEGER NOT NULL DEFAULT 0,
    total_reviews INTEGER NOT NULL DEFAULT 0,
    total_issues INTEGER NOT NULL DEFAULT 0,
    repositories_count INTEGER NOT NULL DEFAULT 0,
    top_repositories JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT check_user_or_team CHECK (
        (user_id IS NOT NULL AND team_id IS NULL) OR 
        (user_id IS NULL AND team_id IS NOT NULL)
    )
);

CREATE INDEX idx_aggregated_metrics_user_id ON aggregated_metrics(user_id);
CREATE INDEX idx_aggregated_metrics_team_id ON aggregated_metrics(team_id);
CREATE INDEX idx_aggregated_metrics_period ON aggregated_metrics(period_start, period_end);
CREATE UNIQUE INDEX uniq_aggregated_metrics_user ON aggregated_metrics(user_id, period_start, period_end) WHERE user_id IS NOT NULL;
CREATE UNIQUE INDEX uniq_aggregated_metrics_team ON aggregated_metrics(team_id, period_start, period_end) WHERE team_id IS NOT NULL;

-- Create configurations table
CREATE TABLE configurations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key VARCHAR(255) NOT NULL UNIQUE,
    value JSONB NOT NULL,
    description TEXT,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL
);

CREATE UNIQUE INDEX idx_configurations_key ON configurations(key);

-- Create sync_history table
CREATE TABLE sync_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider_id UUID REFERENCES vcs_providers(id) ON DELETE SET NULL,
    sync_type VARCHAR(50) NOT NULL CHECK (sync_type IN ('employees', 'git_activities', 'aggregation')),
    status VARCHAR(20) NOT NULL CHECK (status IN ('running', 'completed', 'failed')),
    users_synced INTEGER DEFAULT 0,
    activities_synced INTEGER DEFAULT 0,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT
);

CREATE INDEX idx_sync_history_provider_id ON sync_history(provider_id);
CREATE INDEX idx_sync_history_type ON sync_history(sync_type);
CREATE INDEX idx_sync_history_status ON sync_history(status);
CREATE INDEX idx_sync_history_started_at ON sync_history(started_at DESC);

-- Insert default configurations
INSERT INTO configurations (key, value, description) VALUES
    ('sync_schedule', '"0 */6 * * *"', 'Cron schedule for syncing git activities (every 6 hours)'),
    ('aggregation_schedule', '"0 2 * * *"', 'Cron schedule for aggregating metrics (daily at 2 AM)'),
    ('employee_sync_schedule', '"0 0 * * 0"', 'Cron schedule for syncing employees (weekly on Sunday)'),
    ('default_period_days', '30', 'Default period for analytics in days'),
    ('retention_days_activities', '730', 'Retention period for git_activities in days (2 years)'),
    ('retention_days_sync_history', '180', 'Retention period for sync_history in days (6 months)');

-- Create trigger function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Add triggers for updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_teams_updated_at BEFORE UPDATE ON teams
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_vcs_providers_updated_at BEFORE UPDATE ON vcs_providers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_configurations_updated_at BEFORE UPDATE ON configurations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
