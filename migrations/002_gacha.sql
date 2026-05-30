CREATE TABLE gacha_pool (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(64) NOT NULL,
  cost_item VARCHAR(32) NOT NULL DEFAULT 'diamond',
  cost_one INT UNSIGNED NOT NULL,
  cost_ten INT UNSIGNED NOT NULL,
  pity_limit INT UNSIGNED NOT NULL,
  status TINYINT NOT NULL DEFAULT 1,
  start_at DATETIME NULL,
  end_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE gacha_pool_item (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  pool_id BIGINT UNSIGNED NOT NULL,
  item_type VARCHAR(32) NOT NULL,
  item_id BIGINT UNSIGNED NOT NULL,
  item_count INT UNSIGNED NOT NULL,
  quality TINYINT NOT NULL,
  weight INT UNSIGNED NOT NULL,
  is_pity TINYINT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_pool_id (pool_id),
  KEY idx_pool_pity (pool_id, is_pity, quality)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE player_gacha_state (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  player_id BIGINT UNSIGNED NOT NULL,
  pool_id BIGINT UNSIGNED NOT NULL,
  pity_counter INT UNSIGNED NOT NULL DEFAULT 0,
  total_draw INT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_player_pool (player_id, pool_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE gacha_record (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  player_id BIGINT UNSIGNED NOT NULL,
  pool_id BIGINT UNSIGNED NOT NULL,
  draw_no VARCHAR(64) NOT NULL,
  item_type VARCHAR(32) NOT NULL,
  item_id BIGINT UNSIGNED NOT NULL,
  item_count INT UNSIGNED NOT NULL,
  quality TINYINT NOT NULL,
  is_pity TINYINT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_player_time (player_id, created_at),
  KEY idx_draw_no (draw_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO gacha_pool (id, name, cost_item, cost_one, cost_ten, pity_limit, status)
VALUES
  (1, '天命召唤', 'diamond', 160, 1600, 90, 1);

INSERT INTO gacha_pool_item (pool_id, item_type, item_id, item_count, quality, weight, is_pity)
VALUES
  (1, 'hero', 1, 1, 5, 40, 1),
  (1, 'hero', 2, 1, 4, 120, 0),
  (1, 'hero', 4, 1, 4, 120, 0),
  (1, 'hero', 3, 1, 3, 260, 0),
  (1, 'hero', 5, 1, 3, 260, 0),
  (1, 'gold', 0, 1000, 2, 200, 0);
