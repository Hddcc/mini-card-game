CREATE TABLE player_team (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  player_id BIGINT UNSIGNED NOT NULL,
  slot TINYINT NOT NULL,
  player_hero_id BIGINT UNSIGNED NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_player_slot (player_id, slot),
  UNIQUE KEY uk_player_hero (player_id, player_hero_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE stage_config (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(64) NOT NULL,
  chapter INT UNSIGNED NOT NULL,
  prev_stage_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  stamina_cost INT UNSIGNED NOT NULL,
  recommend_power BIGINT UNSIGNED NOT NULL,
  reward_gold BIGINT UNSIGNED NOT NULL,
  reward_exp INT UNSIGNED NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_chapter (chapter),
  KEY idx_prev_stage (prev_stage_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE player_stage (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  player_id BIGINT UNSIGNED NOT NULL,
  stage_id BIGINT UNSIGNED NOT NULL,
  status TINYINT NOT NULL DEFAULT 0,
  best_power BIGINT UNSIGNED NOT NULL DEFAULT 0,
  first_passed_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_player_stage (player_id, stage_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO stage_config (id, name, chapter, prev_stage_id, stamina_cost, recommend_power, reward_gold, reward_exp)
VALUES
  (1, '花果山试炼', 1, 0, 6, 500, 1000, 20),
  (2, '水帘洞守卫', 1, 1, 6, 900, 1500, 30),
  (3, '东海龙宫', 1, 2, 8, 1300, 2000, 40);