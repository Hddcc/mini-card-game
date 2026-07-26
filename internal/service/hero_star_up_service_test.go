package service

import (
	"errors"
	"testing"

	"mini-card-game/internal/model"
	"mini-card-game/internal/repository"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type heroStarTestEnv struct {
	db            *gorm.DB
	heroRepo      *repository.HeroRepository
	assetRepo     *repository.AssetRepository
	rewardService *RewardService
	heroService   *HeroService
}

func newHeroStarTestEnv(t *testing.T) *heroStarTestEnv {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.HeroConfig{}, &model.PlayerHero{}, &model.PlayerAsset{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	heroRepo := repository.NewHeroRepository(db)
	assetRepo := repository.NewAssetRepository(db)
	return &heroStarTestEnv{
		db:            db,
		heroRepo:      heroRepo,
		assetRepo:     assetRepo,
		rewardService: NewRewardService(assetRepo, heroRepo),
		heroService:   NewHeroService(db, heroRepo, assetRepo),
	}
}

func seedHeroConfig(t *testing.T, db *gorm.DB, id uint64, quality uint8, role string) model.HeroConfig {
	t.Helper()
	cfg := model.HeroConfig{
		ID:          id,
		Name:        "Hero",
		Quality:     quality,
		Role:        role,
		BaseHP:      1000,
		BaseATK:     200,
		BaseDEF:     100,
		PowerFactor: 100,
	}
	if err := db.Create(&cfg).Error; err != nil {
		t.Fatalf("seed hero config: %v", err)
	}
	return cfg
}

func seedPlayerHero(t *testing.T, db *gorm.DB, playerID uint64, configID uint64, star uint32, shard uint32) model.PlayerHero {
	t.Helper()
	hero := model.PlayerHero{
		PlayerID:     playerID,
		HeroConfigID: configID,
		Level:        1,
		Star:         star,
		Shard:        shard,
	}
	if err := db.Create(&hero).Error; err != nil {
		t.Fatalf("seed player hero: %v", err)
	}
	return hero
}

func seedAsset(t *testing.T, db *gorm.DB, playerID uint64, gold uint64) model.PlayerAsset {
	t.Helper()
	asset := model.PlayerAsset{PlayerID: playerID, Gold: gold, Diamond: 0, Stamina: 0}
	if err := db.Create(&asset).Error; err != nil {
		t.Fatalf("seed asset: %v", err)
	}
	return asset
}

func TestDuplicateHeroShardConversionByQuality(t *testing.T) {
	tests := []struct {
		name    string
		quality uint8
		want    uint32
	}{
		{name: "quality 5", quality: 5, want: 3},
		{name: "quality 4", quality: 4, want: 5},
		{name: "quality 3", quality: 3, want: 8},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newHeroStarTestEnv(t)
			playerID := uint64(100 + i)
			configID := uint64(10 + i)
			seedHeroConfig(t, env.db, configID, tt.quality, "warrior")
			hero := seedPlayerHero(t, env.db, playerID, configID, 10, 2)
			seedAsset(t, env.db, playerID, 0)

			var results []GrantResult
			err := env.db.Transaction(func(tx *gorm.DB) error {
				var err error
				results, err = env.rewardService.GrantWithResults(tx, playerID, []model.Reward{{Type: "hero", ID: configID, Count: 1}})
				return err
			})
			if err != nil {
				t.Fatalf("grant duplicate: %v", err)
			}
			if len(results) != 1 || !results[0].IsDuplicate || results[0].ConvertedShards != tt.want {
				t.Fatalf("unexpected grant result: %#v", results)
			}
			var stored model.PlayerHero
			if err := env.db.First(&stored, hero.ID).Error; err != nil {
				t.Fatalf("load hero: %v", err)
			}
			if stored.Star != 10 || stored.Shard != 2+tt.want {
				t.Fatalf("stored hero star=%d shard=%d", stored.Star, stored.Shard)
			}
		})
	}
}

func TestFirstHeroRewardCreatesOwnedHero(t *testing.T) {
	env := newHeroStarTestEnv(t)
	seedHeroConfig(t, env.db, 1, 5, "warrior")
	seedAsset(t, env.db, 1001, 0)

	err := env.db.Transaction(func(tx *gorm.DB) error {
		_, err := env.rewardService.GrantWithResults(tx, 1001, []model.Reward{{Type: "hero", ID: 1, Count: 1}})
		return err
	})
	if err != nil {
		t.Fatalf("grant first hero: %v", err)
	}
	var hero model.PlayerHero
	if err := env.db.Where("player_id = ? AND hero_config_id = ?", 1001, 1).First(&hero).Error; err != nil {
		t.Fatalf("load created hero: %v", err)
	}
	if hero.Level != 1 || hero.Star != 1 || hero.Shard != 0 {
		t.Fatalf("created hero mismatch: %#v", hero)
	}
}

func TestHeroListCreatesStarterHeroWhenInventoryEmpty(t *testing.T) {
	env := newHeroStarTestEnv(t)
	seedHeroConfig(t, env.db, 1, 5, "warrior")

	heroes, err := env.heroService.List(1001)
	if err != nil {
		t.Fatalf("list heroes: %v", err)
	}
	if len(heroes) != 1 {
		t.Fatalf("expected starter hero, got %d heroes", len(heroes))
	}
	if heroes[0].HeroConfigID != 1 || heroes[0].Level != 1 || heroes[0].Star != 1 || heroes[0].Shard != 0 {
		t.Fatalf("starter hero mismatch: %#v", heroes[0])
	}

	heroes, err = env.heroService.List(1001)
	if err != nil {
		t.Fatalf("list heroes again: %v", err)
	}
	if len(heroes) != 1 {
		t.Fatalf("starter hero should not duplicate, got %d heroes", len(heroes))
	}
}

func TestHeroStarUpSuccessAndRejections(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := newHeroStarTestEnv(t)
		seedHeroConfig(t, env.db, 1, 4, "tank")
		hero := seedPlayerHero(t, env.db, 1001, 1, 1, 10)
		seedAsset(t, env.db, 1001, 100)

		out, err := env.heroService.StarUp(1001, hero.ID)
		if err != nil {
			t.Fatalf("star up: %v", err)
		}
		if out.Hero.Star != 2 || out.Hero.Shard != 0 || out.Assets.Gold != 0 {
			t.Fatalf("unexpected output: %#v", out)
		}
	})

	t.Run("unowned", func(t *testing.T) {
		env := newHeroStarTestEnv(t)
		seedHeroConfig(t, env.db, 1, 4, "tank")
		hero := seedPlayerHero(t, env.db, 1001, 1, 1, 10)
		seedAsset(t, env.db, 1002, 100)

		_, err := env.heroService.StarUp(1002, hero.ID)
		if !errors.Is(err, ErrHeroNotOwned) {
			t.Fatalf("expected unowned error, got %v", err)
		}
	})

	t.Run("max star", func(t *testing.T) {
		env := newHeroStarTestEnv(t)
		seedHeroConfig(t, env.db, 1, 4, "tank")
		hero := seedPlayerHero(t, env.db, 1001, 1, 10, 999)
		seedAsset(t, env.db, 1001, 9999)

		_, err := env.heroService.StarUp(1001, hero.ID)
		if !errors.Is(err, ErrHeroMaxStar) {
			t.Fatalf("expected max star error, got %v", err)
		}
	})

	t.Run("insufficient shards", func(t *testing.T) {
		env := newHeroStarTestEnv(t)
		seedHeroConfig(t, env.db, 1, 4, "tank")
		hero := seedPlayerHero(t, env.db, 1001, 1, 1, 9)
		seedAsset(t, env.db, 1001, 100)

		_, err := env.heroService.StarUp(1001, hero.ID)
		if !errors.Is(err, ErrHeroShardNotEnough) {
			t.Fatalf("expected shard error, got %v", err)
		}
	})

	t.Run("insufficient gold", func(t *testing.T) {
		env := newHeroStarTestEnv(t)
		seedHeroConfig(t, env.db, 1, 4, "tank")
		hero := seedPlayerHero(t, env.db, 1001, 1, 1, 10)
		seedAsset(t, env.db, 1001, 99)

		_, err := env.heroService.StarUp(1001, hero.ID)
		if !errors.Is(err, ErrHeroGoldNotEnough) {
			t.Fatalf("expected gold error, got %v", err)
		}
	})
}

func TestHeroBattleStatsUseStarAndRoleGrowth(t *testing.T) {
	warrior := model.HeroConfig{Role: "warrior", BaseHP: 1000, BaseATK: 200, BaseDEF: 100, PowerFactor: 100}
	tank := model.HeroConfig{Role: "tank", BaseHP: 1000, BaseATK: 200, BaseDEF: 100, PowerFactor: 100}

	oneStar := CalcHeroBattleStats(warrior, 1, 1)
	fiveStarWarrior := CalcHeroBattleStats(warrior, 1, 5)
	fiveStarTank := CalcHeroBattleStats(tank, 1, 5)

	if fiveStarWarrior.MaxHP <= oneStar.MaxHP || fiveStarWarrior.ATK <= oneStar.ATK || fiveStarWarrior.DEF <= oneStar.DEF {
		t.Fatalf("higher star should increase all stats: one=%#v five=%#v", oneStar, fiveStarWarrior)
	}
	if fiveStarTank.MaxHP <= fiveStarWarrior.MaxHP {
		t.Fatalf("tank should favor hp growth: tank=%#v warrior=%#v", fiveStarTank, fiveStarWarrior)
	}
	if fiveStarWarrior.ATK <= fiveStarTank.ATK {
		t.Fatalf("warrior should favor atk growth: warrior=%#v tank=%#v", fiveStarWarrior, fiveStarTank)
	}
	if CalcHeroPower(warrior, 1, 5) <= CalcHeroPower(warrior, 1, 1) {
		t.Fatal("higher star should increase power")
	}
}
