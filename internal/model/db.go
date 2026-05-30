package model

import (
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
	); err != nil {
		return err
	}
	if !db.Migrator().HasColumn(&PlayerAsset{}, "stamina_updated_at") {
		if err := db.Migrator().AddColumn(&PlayerAsset{}, "StaminaUpdatedAt"); err != nil {
			return err
		}
	}

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
