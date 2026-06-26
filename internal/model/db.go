package model

import (
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func NewDB(dsn string) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}

func EnsureSchema(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&User{},
		&PlayerProfile{},
		&PlayerAsset{},
		&HeroConfig{},
		&PlayerHero{},
		&GachaPool{},
		&GachaPoolItem{},
		&PlayerGachaState{},
		&GachaRecord{},
		&PlayerTeam{},
		&StageConfig{},
		&PlayerStage{},
		&DailyTaskConfig{},
		&PlayerDailyTask{},
		&EnemyConfig{},
		&StageEnemyConfig{},
		&SkillConfig{},
		&PlayerBattleSession{},
		&CardSkinConfig{},
		&StageEncounterVariant{},
		&StageEncounterEnemy{},
		&ActivityLottery{},
		&ActivityLotteryPrize{},
		&ActivityLotteryRecord{},
		&ActivityLotteryBlacklist{},
		&ActivityLotteryLocalMessage{},
		&ActivityPrizeReleaseState{},
	); err != nil {
		return err
	}
	if !db.Migrator().HasColumn(&PlayerAsset{}, "stamina_updated_at") {
		if err := db.Migrator().AddColumn(&PlayerAsset{}, "StaminaUpdatedAt"); err != nil {
			return err
		}
	}

	if err := seedBaseConfig(db); err != nil {
		return err
	}
	if err := seedBattleConfig(db); err != nil {
		return err
	}
	if err := seedCardBoardConfig(db); err != nil {
		return err
	}
	if err := syncProvidedArtPaths(db); err != nil {
		return err
	}
	return seedActivityLotteryConfig(db)
}

func seedBaseConfig(db *gorm.DB) error {
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create([]HeroConfig{
		{ID: 1, Name: "孙悟空", Quality: 5, Role: "战士", BaseHP: 1200, BaseATK: 260, BaseDEF: 110, PowerFactor: 120},
		{ID: 2, Name: "猪八戒", Quality: 4, Role: "坦克", BaseHP: 1500, BaseATK: 180, BaseDEF: 160, PowerFactor: 110},
		{ID: 3, Name: "沙悟净", Quality: 3, Role: "守护", BaseHP: 1100, BaseATK: 170, BaseDEF: 130, PowerFactor: 100},
		{ID: 4, Name: "小白龙", Quality: 4, Role: "刺客", BaseHP: 980, BaseATK: 240, BaseDEF: 90, PowerFactor: 115},
		{ID: 5, Name: "唐三藏", Quality: 3, Role: "辅助", BaseHP: 900, BaseATK: 130, BaseDEF: 100, PowerFactor: 95},
	}).Error; err != nil {
		return err
	}

	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create([]StageConfig{
		{ID: 1, Name: "花果山试炼", Chapter: 1, PrevStageID: 0, StaminaCost: 6, RecommendPower: 500, RewardGold: 1000, RewardExp: 20},
		{ID: 2, Name: "水帘洞守卫", Chapter: 1, PrevStageID: 1, StaminaCost: 6, RecommendPower: 900, RewardGold: 1500, RewardExp: 30},
		{ID: 3, Name: "东海龙宫", Chapter: 1, PrevStageID: 2, StaminaCost: 8, RecommendPower: 1300, RewardGold: 2000, RewardExp: 40},
	}).Error; err != nil {
		return err
	}

	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create([]DailyTaskConfig{
		{ID: 1, Name: "完成 1 次抽卡", EventType: "gacha_draw", TargetCount: 1, RewardGold: 1000, RewardDiamond: 20, Status: 1},
		{ID: 2, Name: "挑战 1 次关卡", EventType: "stage_fight", TargetCount: 1, RewardGold: 1000, RewardDiamond: 20, Status: 1},
		{ID: 3, Name: "通关 1 次关卡", EventType: "stage_win", TargetCount: 1, RewardGold: 1500, RewardDiamond: 30, Status: 1},
	}).Error; err != nil {
		return err
	}

	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&GachaPool{
		ID: 1, Name: "天命召唤", CostItem: "diamond", CostOne: 160, CostTen: 1600, PityLimit: 90, Status: 1,
	}).Error; err != nil {
		return err
	}

	return db.Clauses(clause.OnConflict{DoNothing: true}).Create([]GachaPoolItem{
		{PoolID: 1, ItemType: "hero", ItemID: 1, ItemCount: 1, Quality: 5, Weight: 40, IsPity: 1},
		{PoolID: 1, ItemType: "hero", ItemID: 2, ItemCount: 1, Quality: 4, Weight: 120},
		{PoolID: 1, ItemType: "hero", ItemID: 4, ItemCount: 1, Quality: 4, Weight: 120},
		{PoolID: 1, ItemType: "hero", ItemID: 3, ItemCount: 1, Quality: 3, Weight: 260},
		{PoolID: 1, ItemType: "hero", ItemID: 5, ItemCount: 1, Quality: 3, Weight: 260},
		{PoolID: 1, ItemType: "gold", ItemID: 0, ItemCount: 1000, Quality: 2, Weight: 200},
	}).Error
}

func seedBattleConfig(db *gorm.DB) error {
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create([]SkillConfig{
		{ID: 1, Code: "warrior_slash", Name: "破魔斩", TargetType: "enemy", EffectType: "damage", Multiplier: 150, CostRage: 50, Cooldown: 1, EffectKey: "effect-slash", AnimationKey: "fx-gold-burst", Description: "对单个敌人造成高额伤害"},
		{ID: 2, Code: "tank_guard", Name: "金身护佑", TargetType: "self", EffectType: "defense_buff", Multiplier: 0, CostRage: 30, Cooldown: 1, DurationRounds: 1, StatDelta: 80, EffectKey: "effect-shield", AnimationKey: "fx-shield", Description: "提升自身防御"},
		{ID: 3, Code: "guardian_heal", Name: "净瓶甘露", TargetType: "ally_lowest", EffectType: "heal", Multiplier: 120, CostRage: 40, Cooldown: 1, EffectKey: "effect-heal", AnimationKey: "fx-water-heal", Description: "治疗生命最低的友方"},
		{ID: 4, Code: "assassin_strike", Name: "疾影突刺", TargetType: "enemy", EffectType: "damage", Multiplier: 170, CostRage: 60, Cooldown: 1, EffectKey: "effect-pierce", AnimationKey: "fx-dragon-sting", Description: "对单个敌人造成爆发伤害"},
		{ID: 5, Code: "support_heal", Name: "佛光回春", TargetType: "ally_lowest", EffectType: "heal", Multiplier: 140, CostRage: 45, Cooldown: 1, EffectKey: "effect-heal", AnimationKey: "fx-buddha-heal", Description: "治疗生命最低的友方"},
		{ID: 6, Code: "warrior_rally", Name: "斗战号令", TargetType: "ally", EffectType: "attack_buff", Multiplier: 0, CostRage: 40, Cooldown: 1, DurationRounds: 2, StatDelta: 60, EffectKey: "effect-rally", AnimationKey: "fx-red-rally", Description: "提升一名友方攻击"},
		{ID: 101, Code: "enemy_bite", Name: "妖袭", TargetType: "ally_lowest", EffectType: "damage", Multiplier: 100, CostRage: 0, Cooldown: 0, EffectKey: "effect-claw", AnimationKey: "fx-bite", Description: "攻击生命最低的英雄"},
		{ID: 102, Code: "enemy_smash", Name: "重击", TargetType: "ally_lowest", EffectType: "damage", Multiplier: 120, CostRage: 0, Cooldown: 1, EffectKey: "effect-smash", AnimationKey: "fx-heavy-smash", Description: "造成更高伤害"},
	}).Error; err != nil {
		return err
	}

	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create([]EnemyConfig{
		{ID: 1, Name: "山猿小妖", Role: "minion", BaseHP: 520, BaseATK: 95, BaseDEF: 35, SkillID: 101, CardArt: "/static/assets/images/enemy-leopard-spirit.png", PortraitArt: "/static/assets/images/enemy-leopard-spirit.png", AttackAnimation: "fx-claw", SkillAnimation: "fx-bite", HitAnimation: "fx-hit-spark", DefeatAnimation: "fx-defeat-smoke", IdleAnimation: "fx-idle-breathe"},
		{ID: 2, Name: "水帘洞守卫", Role: "guard", BaseHP: 780, BaseATK: 125, BaseDEF: 55, SkillID: 102, CardArt: "/static/assets/images/enemy-rhino-spirit.png", PortraitArt: "/static/assets/images/enemy-rhino-spirit.png", AttackAnimation: "fx-smash", SkillAnimation: "fx-heavy-smash", HitAnimation: "fx-hit-spark", DefeatAnimation: "fx-defeat-smoke", IdleAnimation: "fx-idle-breathe"},
		{ID: 3, Name: "东海虾兵", Role: "minion", BaseHP: 620, BaseATK: 115, BaseDEF: 45, SkillID: 101, CardArt: "/static/assets/images/enemy-leopard-spirit.png", PortraitArt: "/static/assets/images/enemy-leopard-spirit.png", AttackAnimation: "fx-spear", SkillAnimation: "fx-bite", HitAnimation: "fx-hit-spark", DefeatAnimation: "fx-defeat-smoke", IdleAnimation: "fx-idle-breathe"},
		{ID: 4, Name: "龙宫巡将", Role: "boss", BaseHP: 1100, BaseATK: 150, BaseDEF: 75, SkillID: 102, CardArt: "/static/assets/images/enemy-rhino-spirit.png", PortraitArt: "/static/assets/images/enemy-rhino-spirit.png", AttackAnimation: "fx-cleave", SkillAnimation: "fx-heavy-smash", HitAnimation: "fx-hit-spark", DefeatAnimation: "fx-defeat-smoke", IdleAnimation: "fx-idle-glow"},
	}).Error; err != nil {
		return err
	}

	return db.Clauses(clause.OnConflict{DoNothing: true}).Create([]StageEnemyConfig{
		{StageID: 1, EnemyConfigID: 1, Slot: 1, Level: 1, Count: 1},
		{StageID: 1, EnemyConfigID: 1, Slot: 2, Level: 1, Count: 1},
		{StageID: 2, EnemyConfigID: 2, Slot: 1, Level: 2, Count: 1},
		{StageID: 2, EnemyConfigID: 1, Slot: 2, Level: 2, Count: 1},
		{StageID: 3, EnemyConfigID: 4, Slot: 1, Level: 3, Count: 1},
		{StageID: 3, EnemyConfigID: 3, Slot: 2, Level: 3, Count: 2},
	}).Error
}

func seedCardBoardConfig(db *gorm.DB) error {
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create([]CardSkinConfig{
		{OwnerType: "hero", OwnerID: 1, CardArt: "/static/assets/images/hero-sun-wukong.png", PortraitArt: "/static/assets/images/hero-sun-wukong.png", AttackAnimation: "fx-slash", SkillAnimation: "fx-gold-burst", HitAnimation: "fx-hit-spark", DefeatAnimation: "fx-defeat-smoke", IdleAnimation: "fx-idle-glow"},
		{OwnerType: "hero", OwnerID: 2, CardArt: "/static/assets/images/hero-zhu-bajie.png", PortraitArt: "/static/assets/images/hero-zhu-bajie.png", AttackAnimation: "fx-smash", SkillAnimation: "fx-shield", HitAnimation: "fx-hit-shield", DefeatAnimation: "fx-defeat-smoke", IdleAnimation: "fx-idle-breathe"},
		{OwnerType: "hero", OwnerID: 3, CardArt: "/static/assets/images/hero-sha-wujing.png", PortraitArt: "/static/assets/images/hero-sha-wujing.png", AttackAnimation: "fx-staff", SkillAnimation: "fx-water-heal", HitAnimation: "fx-hit-spark", DefeatAnimation: "fx-defeat-smoke", IdleAnimation: "fx-idle-breathe"},
		{OwnerType: "hero", OwnerID: 4, CardArt: "/static/assets/images/hero-xiao-bailong.png", PortraitArt: "/static/assets/images/hero-xiao-bailong.png", AttackAnimation: "fx-pierce", SkillAnimation: "fx-dragon-sting", HitAnimation: "fx-hit-spark", DefeatAnimation: "fx-defeat-smoke", IdleAnimation: "fx-idle-glow"},
		{OwnerType: "hero", OwnerID: 5, CardArt: "/static/assets/images/hero-tang-sanzang.png", PortraitArt: "/static/assets/images/hero-tang-sanzang.png", AttackAnimation: "fx-prayer", SkillAnimation: "fx-buddha-heal", HitAnimation: "fx-hit-spark", DefeatAnimation: "fx-defeat-smoke", IdleAnimation: "fx-idle-glow"},
		{OwnerType: "enemy", OwnerID: 1, CardArt: "/static/assets/images/enemy-leopard-spirit.png", PortraitArt: "/static/assets/images/enemy-leopard-spirit.png", AttackAnimation: "fx-claw", SkillAnimation: "fx-bite", HitAnimation: "fx-hit-spark", DefeatAnimation: "fx-defeat-smoke", IdleAnimation: "fx-idle-breathe"},
		{OwnerType: "enemy", OwnerID: 2, CardArt: "/static/assets/images/enemy-rhino-spirit.png", PortraitArt: "/static/assets/images/enemy-rhino-spirit.png", AttackAnimation: "fx-smash", SkillAnimation: "fx-heavy-smash", HitAnimation: "fx-hit-spark", DefeatAnimation: "fx-defeat-smoke", IdleAnimation: "fx-idle-breathe"},
		{OwnerType: "enemy", OwnerID: 3, CardArt: "/static/assets/images/enemy-leopard-spirit.png", PortraitArt: "/static/assets/images/enemy-leopard-spirit.png", AttackAnimation: "fx-spear", SkillAnimation: "fx-bite", HitAnimation: "fx-hit-spark", DefeatAnimation: "fx-defeat-smoke", IdleAnimation: "fx-idle-breathe"},
		{OwnerType: "enemy", OwnerID: 4, CardArt: "/static/assets/images/enemy-rhino-spirit.png", PortraitArt: "/static/assets/images/enemy-rhino-spirit.png", AttackAnimation: "fx-cleave", SkillAnimation: "fx-heavy-smash", HitAnimation: "fx-hit-spark", DefeatAnimation: "fx-defeat-smoke", IdleAnimation: "fx-idle-glow"},
	}).Error; err != nil {
		return err
	}

	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create([]StageEncounterVariant{
		{ID: 1, StageID: 1, Name: "山猿双袭", MinPower: 350, MaxPower: 750, EstimatedPower: 520, Weight: 3, Status: 1},
		{ID: 2, StageID: 1, Name: "山猿突进", MinPower: 350, MaxPower: 750, EstimatedPower: 610, Weight: 2, Status: 1},
		{ID: 3, StageID: 2, Name: "洞口守卫", MinPower: 700, MaxPower: 1150, EstimatedPower: 900, Weight: 3, Status: 1},
		{ID: 4, StageID: 2, Name: "守卫带小妖", MinPower: 700, MaxPower: 1150, EstimatedPower: 1030, Weight: 2, Status: 1},
		{ID: 5, StageID: 3, Name: "龙宫巡阵", MinPower: 1050, MaxPower: 1600, EstimatedPower: 1350, Weight: 3, Status: 1},
		{ID: 6, StageID: 3, Name: "虾兵护将", MinPower: 1050, MaxPower: 1600, EstimatedPower: 1480, Weight: 2, Status: 1},
	}).Error; err != nil {
		return err
	}

	return db.Clauses(clause.OnConflict{DoNothing: true}).Create([]StageEncounterEnemy{
		{ID: 1, VariantID: 1, EnemyConfigID: 1, Slot: 1, Level: 1, Count: 2, SkillID: 101},
		{ID: 2, VariantID: 2, EnemyConfigID: 1, Slot: 1, Level: 2, Count: 3, SkillID: 101},
		{ID: 3, VariantID: 3, EnemyConfigID: 2, Slot: 1, Level: 2, Count: 1, SkillID: 102},
		{ID: 4, VariantID: 3, EnemyConfigID: 1, Slot: 2, Level: 2, Count: 1, SkillID: 101},
		{ID: 5, VariantID: 4, EnemyConfigID: 2, Slot: 1, Level: 3, Count: 1, SkillID: 102},
		{ID: 6, VariantID: 4, EnemyConfigID: 1, Slot: 2, Level: 2, Count: 2, SkillID: 101},
		{ID: 7, VariantID: 5, EnemyConfigID: 4, Slot: 1, Level: 3, Count: 1, SkillID: 102},
		{ID: 8, VariantID: 5, EnemyConfigID: 3, Slot: 2, Level: 3, Count: 1, SkillID: 101},
		{ID: 9, VariantID: 6, EnemyConfigID: 4, Slot: 1, Level: 4, Count: 1, SkillID: 102},
		{ID: 10, VariantID: 6, EnemyConfigID: 3, Slot: 2, Level: 3, Count: 2, SkillID: 101},
	}).Error
}

func seedActivityLotteryConfig(db *gorm.DB) error {
	startAt := time.Date(2025, 1, 1, 0, 0, 0, 0, time.Local)
	endAt := time.Date(2035, 12, 31, 23, 59, 59, 0, time.Local)
	activity := ActivityLottery{
		ID:           1,
		Code:         "peach-festival-blessing",
		Name:         "蟠桃会·天庭赐福",
		Description:  "每日限次抽取天庭赏赐，奖励通过异步发奖到账。",
		BannerImage:  "",
		DailyLimit:   3,
		IPDailyLimit: 20,
		Status:       1,
		StartAt:      startAt,
		EndAt:        endAt,
	}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&activity).Error; err != nil {
		return err
	}

	return db.Clauses(clause.OnConflict{DoNothing: true}).Create([]ActivityLotteryPrize{
		{ID: 1, ActivityID: 1, Name: "灵玉福袋", Description: "灵玉 x88", Icon: "diamond", RewardType: "diamond", RewardCount: 88, Quality: 5, Weight: 80, TotalNum: 80, LeftNum: 80, ReleasePlan: "daily:20", DisplayOrder: 1, Status: 1},
		{ID: 2, ActivityID: 1, Name: "金铢宝匣", Description: "金币 x5000", Icon: "account_balance_wallet", RewardType: "gold", RewardCount: 5000, Quality: 4, Weight: 260, TotalNum: -1, LeftNum: -1, DisplayOrder: 2, Status: 1},
		{ID: 3, ActivityID: 1, Name: "行囊补给", Description: "体力 x30", Icon: "bolt", RewardType: "stamina", RewardCount: 30, Quality: 3, Weight: 300, TotalNum: 300, LeftNum: 300, ReleasePlan: "daily:80", DisplayOrder: 3, Status: 1},
		{ID: 4, ActivityID: 1, Name: "大圣显灵", Description: "孙悟空 x1", Icon: "auto_awesome", RewardType: "hero", RewardID: 1, RewardCount: 1, Quality: 5, Weight: 20, TotalNum: 10, LeftNum: 10, ReleasePlan: "daily:2", DisplayOrder: 4, Status: 1},
		{ID: 5, ActivityID: 1, Name: "安慰香火", Description: "金币 x800", Icon: "redeem", RewardType: "gold", RewardCount: 800, Quality: 2, Weight: 340, TotalNum: -1, LeftNum: -1, DisplayOrder: 5, Fallback: 1, Status: 1},
	}).Error
}

func syncProvidedArtPaths(db *gorm.DB) error {
	if err := db.Model(&ActivityLottery{}).
		Where("code = ? AND banner_image = ?", "peach-festival-blessing", "/static/assets/portraits/tang-sanzang.svg").
		Update("banner_image", "").Error; err != nil {
		return err
	}

	enemyArt := map[uint64]string{
		1: "/static/assets/images/enemy-leopard-spirit.png",
		2: "/static/assets/images/enemy-rhino-spirit.png",
		3: "/static/assets/images/enemy-leopard-spirit.png",
		4: "/static/assets/images/enemy-rhino-spirit.png",
	}
	for id, art := range enemyArt {
		if err := db.Model(&EnemyConfig{}).
			Where("id = ?", id).
			Updates(map[string]interface{}{"card_art": art, "portrait_art": art}).Error; err != nil {
			return err
		}
	}

	skinArt := []CardSkinConfig{
		{OwnerType: "hero", OwnerID: 1, CardArt: "/static/assets/images/hero-sun-wukong.png", PortraitArt: "/static/assets/images/hero-sun-wukong.png"},
		{OwnerType: "hero", OwnerID: 2, CardArt: "/static/assets/images/hero-zhu-bajie.png", PortraitArt: "/static/assets/images/hero-zhu-bajie.png"},
		{OwnerType: "hero", OwnerID: 3, CardArt: "/static/assets/images/hero-sha-wujing.png", PortraitArt: "/static/assets/images/hero-sha-wujing.png"},
		{OwnerType: "hero", OwnerID: 4, CardArt: "/static/assets/images/hero-xiao-bailong.png", PortraitArt: "/static/assets/images/hero-xiao-bailong.png"},
		{OwnerType: "hero", OwnerID: 5, CardArt: "/static/assets/images/hero-tang-sanzang.png", PortraitArt: "/static/assets/images/hero-tang-sanzang.png"},
		{OwnerType: "enemy", OwnerID: 1, CardArt: "/static/assets/images/enemy-leopard-spirit.png", PortraitArt: "/static/assets/images/enemy-leopard-spirit.png"},
		{OwnerType: "enemy", OwnerID: 2, CardArt: "/static/assets/images/enemy-rhino-spirit.png", PortraitArt: "/static/assets/images/enemy-rhino-spirit.png"},
		{OwnerType: "enemy", OwnerID: 3, CardArt: "/static/assets/images/enemy-leopard-spirit.png", PortraitArt: "/static/assets/images/enemy-leopard-spirit.png"},
		{OwnerType: "enemy", OwnerID: 4, CardArt: "/static/assets/images/enemy-rhino-spirit.png", PortraitArt: "/static/assets/images/enemy-rhino-spirit.png"},
	}
	for _, skin := range skinArt {
		if err := db.Model(&CardSkinConfig{}).
			Where("owner_type = ? AND owner_id = ?", skin.OwnerType, skin.OwnerID).
			Updates(map[string]interface{}{"card_art": skin.CardArt, "portrait_art": skin.PortraitArt}).Error; err != nil {
			return err
		}
	}
	return nil
}
