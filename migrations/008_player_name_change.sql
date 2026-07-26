ALTER TABLE player_profiles
  ADD COLUMN name_updated_at DATETIME NULL,
  ADD COLUMN name_change_date DATE NULL,
  ADD COLUMN name_daily_change_count INT UNSIGNED NOT NULL DEFAULT 0;
