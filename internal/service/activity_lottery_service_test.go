package service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"mini-card-game/internal/cache"
	"mini-card-game/internal/model"
	"mini-card-game/internal/repository"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/glebarez/sqlite"
	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

type activityTestEnv struct {
	db        *gorm.DB
	redis     *miniredis.Miniredis
	cache     *cache.RedisActivityCache
	repo      *repository.ActivityLotteryRepository
	publisher *mockActivityPublisher
	svc       *ActivityLotteryService
}

type mockActivityPublisher struct {
	mu        sync.Mutex
	err       error
	published [][]byte
}

func (p *mockActivityPublisher) Publish(ctx context.Context, body []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.err != nil {
		return p.err
	}
	copied := append([]byte(nil), body...)
	p.published = append(p.published, copied)
	return nil
}

func (p *mockActivityPublisher) Count() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.published)
}

func newActivityTestEnv(t *testing.T) *activityTestEnv {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.ActivityLottery{},
		&model.ActivityLotteryPrize{},
		&model.ActivityLotteryRecord{},
		&model.ActivityLotteryBlacklist{},
		&model.ActivityLotteryLocalMessage{},
		&model.ActivityPrizeReleaseState{},
		&model.PlayerAsset{},
		&model.PlayerHero{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mr.Close)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	repo := repository.NewActivityLotteryRepository(db)
	activityCache := cache.NewRedisActivityCache(redisClient)
	publisher := &mockActivityPublisher{}
	svc := NewActivityLotteryService(
		db,
		repo,
		NewRewardService(repository.NewAssetRepository(db), repository.NewHeroRepository(db)),
		activityCache,
		publisher,
	)
	return &activityTestEnv{db: db, redis: mr, cache: activityCache, repo: repo, publisher: publisher, svc: svc}
}

func seedActivityFixture(t *testing.T, env *activityTestEnv, dailyLimit uint32, ipDailyLimit uint32) model.ActivityLottery {
	t.Helper()
	now := time.Now()
	activity := model.ActivityLottery{
		ID:           1,
		Code:         "test-activity",
		Name:         "测试活动",
		Description:  "测试活动",
		DailyLimit:   dailyLimit,
		IPDailyLimit: ipDailyLimit,
		Status:       1,
		StartAt:      now.Add(-time.Hour),
		EndAt:        now.Add(time.Hour),
	}
	if err := env.db.Create(&activity).Error; err != nil {
		t.Fatalf("seed activity: %v", err)
	}
	prizes := []model.ActivityLotteryPrize{
		{ID: 1, ActivityID: activity.ID, Name: "limited", RewardType: "gold", RewardCount: 100, Weight: 100, TotalNum: 1, LeftNum: 1, Status: 1, DisplayOrder: 1},
		{ID: 2, ActivityID: activity.ID, Name: "fallback", RewardType: "gold", RewardCount: 1, Weight: 0, TotalNum: -1, LeftNum: -1, Fallback: 1, Status: 1, DisplayOrder: 2},
	}
	if err := env.db.Create(&prizes).Error; err != nil {
		t.Fatalf("seed prizes: %v", err)
	}
	if err := env.cache.HSetInt(cache.ActivityPrizePoolKey(activity.ID), "1", 1); err != nil {
		t.Fatalf("seed redis prize pool: %v", err)
	}
	return activity
}

func countActivityRecords(t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	var count int64
	if err := db.Model(&model.ActivityLotteryRecord{}).Count(&count).Error; err != nil {
		t.Fatalf("count records: %v", err)
	}
	return count
}

func prizeLeft(t *testing.T, db *gorm.DB, prizeID uint64) int64 {
	t.Helper()
	var prize model.ActivityLotteryPrize
	if err := db.First(&prize, prizeID).Error; err != nil {
		t.Fatalf("load prize: %v", err)
	}
	return prize.LeftNum
}

func TestActivityFallbackPrizePrefersConfiguredFallback(t *testing.T) {
	svc := &ActivityLotteryService{}
	prize, err := svc.fallbackPrize([]model.ActivityLotteryPrize{
		{ID: 1, Name: "limited", TotalNum: 0},
		{ID: 2, Name: "unlimited", TotalNum: -1},
		{ID: 3, Name: "fallback", TotalNum: 0, Fallback: 1},
	})
	if err != nil {
		t.Fatalf("fallback prize: %v", err)
	}
	if prize.ID != 3 {
		t.Fatalf("expected configured fallback, got %d", prize.ID)
	}
}

func TestActivityFallbackPrizeUsesUnlimitedWhenNoFallback(t *testing.T) {
	svc := &ActivityLotteryService{}
	prize, err := svc.fallbackPrize([]model.ActivityLotteryPrize{
		{ID: 1, Name: "limited", TotalNum: 0},
		{ID: 2, Name: "unlimited", TotalNum: -1},
	})
	if err != nil {
		t.Fatalf("fallback prize: %v", err)
	}
	if prize.ID != 2 {
		t.Fatalf("expected unlimited fallback, got %d", prize.ID)
	}
}

func TestActivityUtilityParsing(t *testing.T) {
	if got := parseDailyRelease("daily:20"); got != 20 {
		t.Fatalf("daily release = %d", got)
	}
	if got := parseDailyRelease("bad:20"); got != 0 {
		t.Fatalf("bad release should be 0, got %d", got)
	}
	if got := remaining(3, 1); got != 2 {
		t.Fatalf("remaining = %d", got)
	}
	if got := remaining(3, 5); got != 0 {
		t.Fatalf("overused remaining = %d", got)
	}
}

func TestActivityDrawRejectionsDoNotMutateInventory(t *testing.T) {
	t.Run("daily quota", func(t *testing.T) {
		env := newActivityTestEnv(t)
		activity := seedActivityFixture(t, env, 1, 10)
		if err := env.db.Create(&model.ActivityLotteryRecord{
			ActivityID: activity.ID,
			PrizeID:    2,
			PlayerID:   1001,
			DrawNo:     "daily-used",
			RequestIP:  "10.0.0.1",
			PrizeName:  "fallback",
		}).Error; err != nil {
			t.Fatalf("seed record: %v", err)
		}

		_, err := env.svc.Draw(context.Background(), 1001, "10.0.0.2")
		if err == nil || !strings.Contains(err.Error(), "daily draw limit reached") {
			t.Fatalf("expected daily quota rejection, got %v", err)
		}
		if got := prizeLeft(t, env.db, 1); got != 1 {
			t.Fatalf("limited inventory changed on rejection: %d", got)
		}
		if got := countActivityRecords(t, env.db); got != 1 {
			t.Fatalf("record count changed on rejection: %d", got)
		}
	})

	t.Run("ip limit", func(t *testing.T) {
		env := newActivityTestEnv(t)
		activity := seedActivityFixture(t, env, 10, 1)
		if err := env.db.Create(&model.ActivityLotteryRecord{
			ActivityID: activity.ID,
			PrizeID:    2,
			PlayerID:   1001,
			DrawNo:     "ip-used",
			RequestIP:  "10.0.0.9",
			PrizeName:  "fallback",
		}).Error; err != nil {
			t.Fatalf("seed record: %v", err)
		}

		_, err := env.svc.Draw(context.Background(), 1002, "10.0.0.9")
		if err == nil || !strings.Contains(err.Error(), "ip draw limit reached") {
			t.Fatalf("expected ip quota rejection, got %v", err)
		}
		if got := prizeLeft(t, env.db, 1); got != 1 {
			t.Fatalf("limited inventory changed on rejection: %d", got)
		}
		if got := countActivityRecords(t, env.db); got != 1 {
			t.Fatalf("record count changed on rejection: %d", got)
		}
	})

	t.Run("blacklist", func(t *testing.T) {
		env := newActivityTestEnv(t)
		seedActivityFixture(t, env, 10, 10)
		if err := env.db.Create(&model.ActivityLotteryBlacklist{
			TargetType: "player",
			Target:     "1001",
			Status:     1,
		}).Error; err != nil {
			t.Fatalf("seed blacklist: %v", err)
		}

		_, err := env.svc.Draw(context.Background(), 1001, "10.0.0.3")
		if err == nil || !strings.Contains(err.Error(), "player blacklisted") {
			t.Fatalf("expected blacklist rejection, got %v", err)
		}
		if got := prizeLeft(t, env.db, 1); got != 1 {
			t.Fatalf("limited inventory changed on rejection: %d", got)
		}
		if got := countActivityRecords(t, env.db); got != 0 {
			t.Fatalf("record count changed on rejection: %d", got)
		}
	})
}

func TestActivityPlayerDrawLockRejectsDuplicateClickAndReleasesOnlyOwner(t *testing.T) {
	env := newActivityTestEnv(t)
	seedActivityFixture(t, env, 10, 10)
	lockKey := cache.ActivityPlayerDrawLockKey(1001)
	locked, err := env.cache.AcquireLock(lockKey, "owner-a", time.Minute)
	if err != nil || !locked {
		t.Fatalf("seed draw lock locked=%v err=%v", locked, err)
	}

	_, err = env.svc.Draw(context.Background(), 1001, "10.0.0.4")
	if err == nil || !strings.Contains(err.Error(), "draw request is processing") {
		t.Fatalf("expected duplicate-click rejection, got %v", err)
	}
	if got := countActivityRecords(t, env.db); got != 0 {
		t.Fatalf("duplicate click created records: %d", got)
	}

	if err := env.cache.ReleaseLock(lockKey, "owner-b"); err != nil {
		t.Fatalf("release wrong token: %v", err)
	}
	locked, err = env.cache.AcquireLock(lockKey, "owner-c", time.Minute)
	if err != nil {
		t.Fatalf("reacquire after wrong release: %v", err)
	}
	if locked {
		t.Fatal("lock was released by a non-owner token")
	}
	if err := env.cache.ReleaseLock(lockKey, "owner-a"); err != nil {
		t.Fatalf("release owner token: %v", err)
	}
	locked, err = env.cache.AcquireLock(lockKey, "owner-c", time.Minute)
	if err != nil || !locked {
		t.Fatalf("owner release did not free lock locked=%v err=%v", locked, err)
	}
}

func TestActivityLimitedInventoryConditionalDecrementPreventsOversell(t *testing.T) {
	env := newActivityTestEnv(t)
	seedActivityFixture(t, env, 10, 10)

	ok, err := env.repo.DecrementPrizeInventory(env.db, 1, 1)
	if err != nil || !ok {
		t.Fatalf("first decrement ok=%v err=%v", ok, err)
	}
	ok, err = env.repo.DecrementPrizeInventory(env.db, 1, 1)
	if err != nil {
		t.Fatalf("second decrement err=%v", err)
	}
	if ok {
		t.Fatal("second decrement succeeded with no inventory left")
	}
	if got := prizeLeft(t, env.db, 1); got != 0 {
		t.Fatalf("inventory oversold or drifted: %d", got)
	}
}

func TestActivityDrawFallsBackWhenRedisPoolEmpty(t *testing.T) {
	env := newActivityTestEnv(t)
	seedActivityFixture(t, env, 10, 10)
	if err := env.cache.HSetInt(cache.ActivityPrizePoolKey(1), "1", 0); err != nil {
		t.Fatalf("empty redis pool: %v", err)
	}

	out, err := env.svc.Draw(context.Background(), 1001, "10.0.0.5")
	if err != nil {
		t.Fatalf("draw fallback: %v", err)
	}
	if out.Prize.ID != 2 {
		t.Fatalf("expected fallback prize, got %d", out.Prize.ID)
	}
	if got := prizeLeft(t, env.db, 1); got != 1 {
		t.Fatalf("limited inventory changed when redis stock unavailable: %d", got)
	}
}

func TestActivityRabbitMQPublishFailureAndCompensation(t *testing.T) {
	env := newActivityTestEnv(t)
	seedActivityFixture(t, env, 10, 10)
	env.publisher.err = errors.New("rabbitmq down")

	out, err := env.svc.Draw(context.Background(), 1001, "10.0.0.6")
	if err != nil {
		t.Fatalf("draw with failed publisher should keep local message: %v", err)
	}
	var message model.ActivityLotteryLocalMessage
	if err := env.db.Where("business_id = ?", out.DrawNo).First(&message).Error; err != nil {
		t.Fatalf("load local message: %v", err)
	}
	if message.Status != MessageFailed || message.RetryCount != 1 {
		t.Fatalf("expected failed retryable message, status=%d retry=%d", message.Status, message.RetryCount)
	}

	env.publisher.err = nil
	past := time.Now().Add(-time.Second)
	message.NextRetryAt = &past
	if err := env.db.Save(&message).Error; err != nil {
		t.Fatalf("make retryable: %v", err)
	}
	if err := env.svc.RetryPendingMessages(context.Background(), 10, 3); err != nil {
		t.Fatalf("retry pending messages: %v", err)
	}
	if env.publisher.Count() != 1 {
		t.Fatalf("expected one compensation publish, got %d", env.publisher.Count())
	}
}

func TestActivityScheduledJobLocksAndPrizePoolRefreshIdempotency(t *testing.T) {
	env := newActivityTestEnv(t)
	now := time.Now()
	activity := model.ActivityLottery{
		ID:           1,
		Code:         "refresh",
		Name:         "refresh",
		DailyLimit:   10,
		IPDailyLimit: 10,
		Status:       1,
		StartAt:      now.Add(-time.Hour),
		EndAt:        now.Add(time.Hour),
	}
	if err := env.db.Create(&activity).Error; err != nil {
		t.Fatalf("seed activity: %v", err)
	}
	prize := model.ActivityLotteryPrize{
		ID:          10,
		ActivityID:  activity.ID,
		Name:        "daily limited",
		RewardType:  "diamond",
		RewardCount: 1,
		Weight:      1,
		TotalNum:    10,
		LeftNum:     10,
		ReleasePlan: "daily:3",
		Status:      1,
	}
	if err := env.db.Create(&prize).Error; err != nil {
		t.Fatalf("seed prize: %v", err)
	}

	if err := env.svc.RefreshPrizePool(context.Background()); err != nil {
		t.Fatalf("first refresh: %v", err)
	}
	value, ok, err := env.cache.HGetInt(cache.ActivityPrizePoolKey(activity.ID), "10")
	if err != nil || !ok || value != 3 {
		t.Fatalf("first refresh stock value=%d ok=%v err=%v", value, ok, err)
	}
	if err := env.svc.RefreshPrizePool(context.Background()); err != nil {
		t.Fatalf("second refresh: %v", err)
	}
	value, ok, err = env.cache.HGetInt(cache.ActivityPrizePoolKey(activity.ID), "10")
	if err != nil || !ok || value != 3 {
		t.Fatalf("second refresh should be idempotent value=%d ok=%v err=%v", value, ok, err)
	}

	locked, err := env.cache.AcquireLock(cache.ActivityPrizePoolRefreshLockKey(), "other-owner", time.Minute)
	if err != nil || !locked {
		t.Fatalf("seed refresh lock locked=%v err=%v", locked, err)
	}
	if err := env.svc.RefreshPrizePool(context.Background()); err != nil {
		t.Fatalf("locked refresh should exit cleanly: %v", err)
	}
	value, _, _ = env.cache.HGetInt(cache.ActivityPrizePoolKey(activity.ID), "10")
	if value != 3 {
		t.Fatalf("locked refresh mutated stock: %d", value)
	}
}
