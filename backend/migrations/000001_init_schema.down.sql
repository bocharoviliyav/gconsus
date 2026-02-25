-- Drop triggers
DROP TRIGGER IF EXISTS update_configurations_updated_at ON configurations;
DROP TRIGGER IF EXISTS update_vcs_providers_updated_at ON vcs_providers;
DROP TRIGGER IF EXISTS update_teams_updated_at ON teams;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse order
DROP TABLE IF EXISTS sync_history;
DROP TABLE IF EXISTS configurations;
DROP TABLE IF EXISTS aggregated_metrics;
DROP TABLE IF EXISTS git_activities;
DROP TABLE IF EXISTS vcs_providers;
DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS users;

-- Drop extension
DROP EXTENSION IF EXISTS "uuid-ossp";
