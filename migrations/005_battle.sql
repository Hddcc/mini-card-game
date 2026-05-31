CREATE TABLE enemy_config (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(64) NOT NULL,
  role VARCHAR(32) NOT NULL,
  base_hp INT UNSIGNED NOT NULL,
  base_atk INT UNSIGNED NOT NULL,
  base_def INT UNSIGNED NOT NULL,
  skill_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE stage_enemy_config (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  stage_id BIGINT UNSIGNED NOT NULL,
  enemy_config_id BIGINT UNSIGNED NOT NULL,
  slot TINYINT NOT NULL,
  level INT UNSIGNED NOT NULL DEFAULT 1,
  count INT UNSIGNED NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_stage_id (stage_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE skill_config (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  code VARCHAR(64) NOT NULL,
  name VARCHAR(64) NOT NULL,
  target_type VARCHAR(32) NOT NULL,
  effect_type VARCHAR(32) NOT NULL,
  multiplier INT UNSIGNED NOT NULL,
  cost_rage INT UNSIGNED NOT NULL DEFAULT 0,
  cooldown INT UNSIGNED NOT NULL DEFAULT 0,
  description VARCHAR(255) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE player_battle_session (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  player_id BIGINT UNSIGNED NOT NULL,
  stage_id BIGINT UNSIGNED NOT NULL,
  status VARCHAR(24) NOT NULL,
  round INT UNSIGNED NOT NULL DEFAULT 1,
  state_json LONGTEXT NOT NULL,
  result_json LONGTEXT NULL,
  expires_at DATETIME NOT NULL,
  settled_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_player_status (player_id, status),
  KEY idx_stage_id (stage_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO skill_config (id, code, name, target_type, effect_type, multiplier, cost_rage, cooldown, description)
VALUES
  (1, 'warrior_slash', '破魔斩', 'enemy', 'damage', 150, 50, 1, '对单个敌人造成高额伤害'),
  (2, 'tank_guard', '金身护佑', 'self', 'defend', 0, 30, 1, '进入防御姿态并获得怒气'),
  (3, 'guardian_heal', '净瓶甘露', 'ally_lowest', 'heal', 120, 40, 1, '治疗生命最低的友方'),
  (4, 'assassin_strike', '疾影突刺', 'enemy', 'damage', 170, 60, 1, '对单个敌人造成爆发伤害'),
  (5, 'support_heal', '佛光回春', 'ally_lowest', 'heal', 140, 45, 1, '治疗生命最低的友方'),
  (101, 'enemy_bite', '妖袭', 'ally_lowest', 'damage', 100, 0, 0, '攻击生命最低的英雄'),
  (102, 'enemy_smash', '重击', 'ally_lowest', 'damage', 120, 0, 1, '造成更高伤害');

INSERT INTO enemy_config (id, name, role, base_hp, base_atk, base_def, skill_id)
VALUES
  (1, '山猿小妖', 'minion', 520, 95, 35, 101),
  (2, '水帘洞守卫', 'guard', 780, 125, 55, 102),
  (3, '东海虾兵', 'minion', 620, 115, 45, 101),
  (4, '龙宫巡将', 'boss', 1100, 150, 75, 102);

INSERT INTO stage_enemy_config (stage_id, enemy_config_id, slot, level, count)
VALUES
  (1, 1, 1, 1, 1),
  (1, 1, 2, 1, 1),
  (2, 2, 1, 2, 1),
  (2, 1, 2, 2, 1),
  (3, 4, 1, 3, 1),
  (3, 3, 2, 3, 2);
