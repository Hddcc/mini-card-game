CREATE TABLE daily_task_config (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(64) NOT NULL,
  event_type VARCHAR(32) NOT NULL,
  target_count INT UNSIGNED NOT NULL,
  reward_gold BIGINT UNSIGNED NOT NULL DEFAULT 0,
  reward_diamond BIGINT UNSIGNED NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_event_status (event_type, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE player_daily_task (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  player_id BIGINT UNSIGNED NOT NULL,
  task_id BIGINT UNSIGNED NOT NULL,
  task_date DATE NOT NULL,
  progress INT UNSIGNED NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 0,
  claimed_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_player_task_date (player_id, task_id, task_date),
  KEY idx_player_date (player_id, task_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO daily_task_config (id, name, event_type, target_count, reward_gold, reward_diamond, status)
VALUES
  (1, '完成 1 次抽卡', 'gacha_draw', 1, 1000, 20, 1),
  (2, '挑战 1 次关卡', 'stage_fight', 1, 1000, 20, 1),
  (3, '通关 1 次关卡', 'stage_win', 1, 1500, 30, 1);

  