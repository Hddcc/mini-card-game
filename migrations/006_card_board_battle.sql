ALTER TABLE enemy_config
  ADD COLUMN card_art VARCHAR(255) NOT NULL DEFAULT '',
  ADD COLUMN portrait_art VARCHAR(255) NOT NULL DEFAULT '',
  ADD COLUMN attack_animation VARCHAR(64) NOT NULL DEFAULT '',
  ADD COLUMN skill_animation VARCHAR(64) NOT NULL DEFAULT '',
  ADD COLUMN hit_animation VARCHAR(64) NOT NULL DEFAULT '',
  ADD COLUMN defeat_animation VARCHAR(64) NOT NULL DEFAULT '',
  ADD COLUMN idle_animation VARCHAR(64) NOT NULL DEFAULT '';

ALTER TABLE skill_config
  ADD COLUMN duration_rounds INT UNSIGNED NOT NULL DEFAULT 0,
  ADD COLUMN stat_delta INT NOT NULL DEFAULT 0,
  ADD COLUMN effect_key VARCHAR(64) NOT NULL DEFAULT '',
  ADD COLUMN animation_key VARCHAR(64) NOT NULL DEFAULT '';

CREATE TABLE card_skin_config (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  owner_type VARCHAR(16) NOT NULL,
  owner_id BIGINT UNSIGNED NOT NULL,
  card_art VARCHAR(255) NOT NULL DEFAULT '',
  portrait_art VARCHAR(255) NOT NULL DEFAULT '',
  attack_animation VARCHAR(64) NOT NULL DEFAULT '',
  skill_animation VARCHAR(64) NOT NULL DEFAULT '',
  hit_animation VARCHAR(64) NOT NULL DEFAULT '',
  defeat_animation VARCHAR(64) NOT NULL DEFAULT '',
  idle_animation VARCHAR(64) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_owner (owner_type, owner_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE stage_encounter_variant (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  stage_id BIGINT UNSIGNED NOT NULL,
  name VARCHAR(64) NOT NULL,
  min_power BIGINT UNSIGNED NOT NULL DEFAULT 0,
  max_power BIGINT UNSIGNED NOT NULL DEFAULT 0,
  estimated_power BIGINT UNSIGNED NOT NULL DEFAULT 0,
  weight INT UNSIGNED NOT NULL DEFAULT 1,
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_stage_status (stage_id, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE stage_encounter_enemy (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  variant_id BIGINT UNSIGNED NOT NULL,
  enemy_config_id BIGINT UNSIGNED NOT NULL,
  slot TINYINT NOT NULL,
  level INT UNSIGNED NOT NULL DEFAULT 1,
  count INT UNSIGNED NOT NULL DEFAULT 1,
  skill_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_variant_id (variant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO card_skin_config (owner_type, owner_id, card_art, portrait_art, attack_animation, skill_animation, hit_animation, defeat_animation, idle_animation)
VALUES
  ('hero', 1, '/static/assets/images/hero-sun-wukong.png', '/static/assets/images/hero-sun-wukong.png', 'fx-slash', 'fx-gold-burst', 'fx-hit-spark', 'fx-defeat-smoke', 'fx-idle-glow'),
  ('hero', 2, '/static/assets/images/hero-zhu-bajie.png', '/static/assets/images/hero-zhu-bajie.png', 'fx-smash', 'fx-shield', 'fx-hit-shield', 'fx-defeat-smoke', 'fx-idle-breathe'),
  ('hero', 3, '/static/assets/images/hero-sha-wujing.png', '/static/assets/images/hero-sha-wujing.png', 'fx-staff', 'fx-water-heal', 'fx-hit-spark', 'fx-defeat-smoke', 'fx-idle-breathe'),
  ('hero', 4, '/static/assets/images/hero-xiao-bailong.png', '/static/assets/images/hero-xiao-bailong.png', 'fx-pierce', 'fx-dragon-sting', 'fx-hit-spark', 'fx-defeat-smoke', 'fx-idle-glow'),
  ('hero', 5, '/static/assets/images/hero-tang-sanzang.png', '/static/assets/images/hero-tang-sanzang.png', 'fx-prayer', 'fx-buddha-heal', 'fx-hit-spark', 'fx-defeat-smoke', 'fx-idle-glow'),
  ('enemy', 1, '/static/assets/images/enemy-leopard-spirit.png', '/static/assets/images/enemy-leopard-spirit.png', 'fx-claw', 'fx-bite', 'fx-hit-spark', 'fx-defeat-smoke', 'fx-idle-breathe'),
  ('enemy', 2, '/static/assets/images/enemy-rhino-spirit.png', '/static/assets/images/enemy-rhino-spirit.png', 'fx-smash', 'fx-heavy-smash', 'fx-hit-spark', 'fx-defeat-smoke', 'fx-idle-breathe'),
  ('enemy', 3, '/static/assets/images/enemy-leopard-spirit.png', '/static/assets/images/enemy-leopard-spirit.png', 'fx-spear', 'fx-bite', 'fx-hit-spark', 'fx-defeat-smoke', 'fx-idle-breathe'),
  ('enemy', 4, '/static/assets/images/enemy-rhino-spirit.png', '/static/assets/images/enemy-rhino-spirit.png', 'fx-cleave', 'fx-heavy-smash', 'fx-hit-spark', 'fx-defeat-smoke', 'fx-idle-glow');

INSERT INTO stage_encounter_variant (id, stage_id, name, min_power, max_power, estimated_power, weight, status)
VALUES
  (1, 1, '山猿双袭', 350, 750, 520, 3, 1),
  (2, 1, '山猿突进', 350, 750, 610, 2, 1),
  (3, 2, '洞口守卫', 700, 1150, 900, 3, 1),
  (4, 2, '守卫带小妖', 700, 1150, 1030, 2, 1),
  (5, 3, '龙宫巡阵', 1050, 1600, 1350, 3, 1),
  (6, 3, '虾兵护将', 1050, 1600, 1480, 2, 1);

INSERT INTO stage_encounter_enemy (id, variant_id, enemy_config_id, slot, level, count, skill_id)
VALUES
  (1, 1, 1, 1, 1, 2, 101),
  (2, 2, 1, 1, 2, 3, 101),
  (3, 3, 2, 1, 2, 1, 102),
  (4, 3, 1, 2, 2, 1, 101),
  (5, 4, 2, 1, 3, 1, 102),
  (6, 4, 1, 2, 2, 2, 101),
  (7, 5, 4, 1, 3, 1, 102),
  (8, 5, 3, 2, 3, 1, 101),
  (9, 6, 4, 1, 4, 1, 102),
  (10, 6, 3, 2, 3, 2, 101);
