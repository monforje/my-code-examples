-- +goose Up
CREATE UNIQUE INDEX IF NOT EXISTS uq_user_profiles_display_name ON user_profiles (display_name) WHERE deleted_at IS NULL AND display_name != '';

-- +goose Down
DROP INDEX IF EXISTS uq_user_profiles_display_name;
