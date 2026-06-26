package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"mini-card-game/internal/cache"
	"mini-card-game/internal/model"
	"mini-card-game/internal/pkg/random"
	"mini-card-game/internal/repository"

	"gorm.io/gorm"
)

const (
	DeliveryPending uint8 = 0
	DeliverySuccess uint8 = 1
	DeliveryFailed  uint8 = 2

	MessagePending   uint8 = 0
	MessageDelivered uint8 = 1
	MessageFailed    uint8 = 2
	MessageExhausted uint8 = 3

	messageTypeAward = "activity_award"
)

type ActivityAwardPublisher interface {
	Publish(ctx context.Context, body []byte) error
}

type ActivityLotteryService struct {
	db            *gorm.DB
	repo          *repository.ActivityLotteryRepository
	rewardService *RewardService
	activityCache *cache.RedisActivityCache
	publisher     ActivityAwardPublisher
}

type ActivityStateView struct {
	Active   bool                 `json:"active"`
	Activity *ActivityView        `json:"activity,omitempty"`
	Quota    *ActivityQuotaView   `json:"quota,omitempty"`
	Eligible bool                 `json:"eligible"`
	Reason   string               `json:"reason,omitempty"`
	Prizes   []ActivityPrizeView  `json:"prizes"`
	Records  []ActivityRecordView `json:"records"`
}

type ActivityView struct {
	ID          uint64 `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	BannerImage string `json:"banner_image"`
	StartAt     string `json:"start_at"`
	EndAt       string `json:"end_at"`
}

type ActivityQuotaView struct {
	DailyLimit     uint32 `json:"daily_limit"`
	UsedToday      uint32 `json:"used_today"`
	RemainingToday uint32 `json:"remaining_today"`
	IPDailyLimit   uint32 `json:"ip_daily_limit"`
	IPUsedToday    uint32 `json:"ip_used_today"`
}

type ActivityPrizeView struct {
	ID           uint64 `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Icon         string `json:"icon"`
	RewardType   string `json:"reward_type"`
	RewardID     uint64 `json:"reward_id"`
	RewardCount  uint64 `json:"reward_count"`
	Quality      uint8  `json:"quality"`
	LeftNum      int64  `json:"left_num"`
	Unlimited    bool   `json:"unlimited"`
	DisplayOrder uint32 `json:"display_order"`
}

type ActivityRecordView struct {
	DrawNo         string `json:"draw_no"`
	PrizeName      string `json:"prize_name"`
	RewardType     string `json:"reward_type"`
	RewardID       uint64 `json:"reward_id"`
	RewardCount    uint64 `json:"reward_count"`
	DeliveryStatus uint8  `json:"delivery_status"`
	CreatedAt      string `json:"created_at"`
}

type ActivityDrawOutput struct {
	DrawNo         string            `json:"draw_no"`
	Prize          ActivityPrizeView `json:"prize"`
	Quota          ActivityQuotaView `json:"quota"`
	DeliveryStatus uint8             `json:"delivery_status"`
}

type ActivityAwardMessage struct {
	BusinessID  string       `json:"business_id"`
	DrawNo      string       `json:"draw_no"`
	DrawID      uint64       `json:"draw_id"`
	PlayerID    uint64       `json:"player_id"`
	ActivityID  uint64       `json:"activity_id"`
	PrizeID     uint64       `json:"prize_id"`
	Reward      model.Reward `json:"reward"`
	RequestedAt string       `json:"requested_at"`
}

func NewActivityLotteryService(db *gorm.DB, repo *repository.ActivityLotteryRepository, rewardService *RewardService, activityCache *cache.RedisActivityCache, publisher ActivityAwardPublisher) *ActivityLotteryService {
	return &ActivityLotteryService{
		db:            db,
		repo:          repo,
		rewardService: rewardService,
		activityCache: activityCache,
		publisher:     publisher,
	}
}

func (s *ActivityLotteryService) State(playerID uint64, ip string) (*ActivityStateView, error) {
	now := time.Now()
	activity, err := s.repo.FindActive(now)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &ActivityStateView{Active: false, Eligible: false, Reason: "no active activity"}, nil
		}
		return nil, err
	}
	prizes, err := s.activePrizes(activity.ID)
	if err != nil {
		return nil, err
	}
	quota, eligible, reason, err := s.quotaAndEligibility(playerID, ip, activity, now)
	if err != nil {
		return nil, err
	}
	records, err := s.repo.ListPlayerRecords(playerID, 10)
	if err != nil {
		return nil, err
	}

	return &ActivityStateView{
		Active:   true,
		Activity: s.activityView(activity),
		Quota:    quota,
		Eligible: eligible,
		Reason:   reason,
		Prizes:   s.prizeViews(activity.ID, prizes),
		Records:  s.recordViews(records),
	}, nil
}

func (s *ActivityLotteryService) Draw(ctx context.Context, playerID uint64, ip string) (*ActivityDrawOutput, error) {
	now := time.Now()
	activity, err := s.repo.FindActive(now)
	if err != nil {
		return nil, errors.New("no active activity")
	}

	lockKey := cache.ActivityPlayerDrawLockKey(playerID)
	lockToken := fmt.Sprintf("%d-%d", playerID, time.Now().UnixNano())
	locked, err := s.activityCache.AcquireLock(lockKey, lockToken, 5*time.Second)
	if err != nil {
		return nil, err
	}
	if !locked {
		return nil, errors.New("draw request is processing")
	}
	defer s.activityCache.ReleaseLock(lockKey, lockToken)

	quota, eligible, reason, err := s.quotaAndEligibility(playerID, ip, activity, now)
	if err != nil {
		return nil, err
	}
	if !eligible {
		return nil, errors.New(reason)
	}

	prizes, err := s.activePrizes(activity.ID)
	if err != nil {
		return nil, err
	}
	prize, point, err := s.pickPrize(activity.ID, prizes)
	if err != nil {
		return nil, err
	}
	reservedPrizeID := uint64(0)

	if prize.TotalNum >= 0 {
		ok, err := s.reserveRedisPrize(activity.ID, prize.ID)
		if err != nil {
			return nil, err
		}
		if !ok {
			fallback, fallbackErr := s.fallbackPrize(prizes)
			if fallbackErr != nil {
				return nil, fallbackErr
			}
			prize = fallback
		} else {
			reservedPrizeID = prize.ID
		}
	}

	var record *model.ActivityLotteryRecord
	var message *model.ActivityLotteryLocalMessage
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if prize.TotalNum >= 0 {
			ok, err := s.repo.DecrementPrizeInventory(tx, prize.ID, 1)
			if err != nil {
				return err
			}
			if !ok {
				if reservedPrizeID > 0 {
					_, _ = s.activityCache.HIncrBy(cache.ActivityPrizePoolKey(activity.ID), strconv.FormatUint(reservedPrizeID, 10), 1)
					reservedPrizeID = 0
				}
				fallback, fallbackErr := s.fallbackPrize(prizes)
				if fallbackErr != nil {
					return fallbackErr
				}
				prize = fallback
			}
		}

		drawNo := fmt.Sprintf("A%d%d", time.Now().UnixNano(), playerID)
		record = &model.ActivityLotteryRecord{
			ActivityID:     activity.ID,
			PrizeID:        prize.ID,
			PlayerID:       playerID,
			DrawNo:         drawNo,
			RandomPoint:    point,
			PrizeName:      prize.Name,
			RewardType:     prize.RewardType,
			RewardID:       prize.RewardID,
			RewardCount:    prize.RewardCount,
			DeliveryStatus: DeliveryPending,
			RequestIP:      ip,
		}
		if err := s.repo.CreateRecord(tx, record); err != nil {
			return err
		}

		award := ActivityAwardMessage{
			BusinessID:  drawNo,
			DrawNo:      drawNo,
			DrawID:      record.ID,
			PlayerID:    playerID,
			ActivityID:  activity.ID,
			PrizeID:     prize.ID,
			Reward:      model.Reward{Type: prize.RewardType, ID: prize.RewardID, Count: prize.RewardCount},
			RequestedAt: time.Now().Format(time.RFC3339),
		}
		payload, err := json.Marshal(award)
		if err != nil {
			return err
		}
		nextRetry := time.Now().Add(10 * time.Second)
		message = &model.ActivityLotteryLocalMessage{
			BusinessID:  drawNo,
			MessageType: messageTypeAward,
			Payload:     string(payload),
			Status:      MessagePending,
			NextRetryAt: &nextRetry,
		}
		return s.repo.CreateLocalMessage(tx, message)
	})
	if err != nil {
		if reservedPrizeID > 0 {
			_, _ = s.activityCache.HIncrBy(cache.ActivityPrizePoolKey(activity.ID), strconv.FormatUint(reservedPrizeID, 10), 1)
		}
		return nil, err
	}

	if message != nil {
		_ = s.PublishLocalMessage(ctx, message)
	}
	s.bumpCounters(activity, playerID, ip)
	quota.UsedToday++
	quota.IPUsedToday++
	quota.RemainingToday = remaining(activity.DailyLimit, quota.UsedToday)

	return &ActivityDrawOutput{
		DrawNo:         record.DrawNo,
		Prize:          s.prizeView(activity.ID, prize),
		Quota:          *quota,
		DeliveryStatus: DeliveryPending,
	}, nil
}

func (s *ActivityLotteryService) Records(playerID uint64) ([]ActivityRecordView, error) {
	records, err := s.repo.ListPlayerRecords(playerID, 20)
	if err != nil {
		return nil, err
	}
	return s.recordViews(records), nil
}

func (s *ActivityLotteryService) PublishLocalMessage(ctx context.Context, message *model.ActivityLotteryLocalMessage) error {
	if s.publisher == nil {
		return errors.New("rabbitmq publisher unavailable")
	}
	err := s.publisher.Publish(ctx, []byte(message.Payload))
	if err != nil {
		s.markMessageFailed(message.BusinessID, err.Error())
		return err
	}
	return nil
}

func (s *ActivityLotteryService) ConsumeAward(payload []byte) error {
	var award ActivityAwardMessage
	if err := json.Unmarshal(payload, &award); err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		message, err := s.repo.FindLocalMessageForUpdate(tx, award.BusinessID)
		if err != nil {
			return err
		}
		if message.Status == MessageDelivered {
			return nil
		}
		record, err := s.repo.FindRecordForUpdate(tx, award.DrawNo)
		if err != nil {
			return err
		}
		if record.DeliveryStatus == DeliverySuccess {
			message.Status = MessageDelivered
			return s.repo.SaveLocalMessage(tx, message)
		}
		if err := s.rewardService.Grant(tx, award.PlayerID, []model.Reward{award.Reward}); err != nil {
			message.Status = MessageFailed
			message.RetryCount++
			message.LastError = err.Error()
			next := time.Now().Add(time.Duration(message.RetryCount+1) * time.Minute)
			message.NextRetryAt = &next
			_ = s.repo.SaveLocalMessage(tx, message)
			return err
		}
		message.Status = MessageDelivered
		message.LastError = ""
		record.DeliveryStatus = DeliverySuccess
		if err := s.repo.SaveLocalMessage(tx, message); err != nil {
			return err
		}
		return s.repo.SaveRecord(tx, record)
	})
}

func (s *ActivityLotteryService) RetryPendingMessages(ctx context.Context, limit int, maxRetry uint32) error {
	token := fmt.Sprintf("message-scan-%d", time.Now().UnixNano())
	locked, err := s.activityCache.AcquireLock(cache.ActivityMessageScanLockKey(), token, 30*time.Second)
	if err != nil || !locked {
		return err
	}
	defer s.activityCache.ReleaseLock(cache.ActivityMessageScanLockKey(), token)

	messages, err := s.repo.ListRetryableMessages(time.Now(), limit)
	if err != nil {
		return err
	}
	for _, message := range messages {
		if message.RetryCount >= maxRetry {
			_ = s.db.Transaction(func(tx *gorm.DB) error {
				lockedMessage, err := s.repo.FindLocalMessageForUpdate(tx, message.BusinessID)
				if err != nil {
					return err
				}
				lockedMessage.Status = MessageExhausted
				return s.repo.SaveLocalMessage(tx, lockedMessage)
			})
			continue
		}
		_ = s.PublishLocalMessage(ctx, &message)
	}
	return nil
}

func (s *ActivityLotteryService) RefreshPrizePool(ctx context.Context) error {
	if !s.activityCache.Available() {
		return nil
	}
	token := fmt.Sprintf("pool-refresh-%d", time.Now().UnixNano())
	locked, err := s.activityCache.AcquireLock(cache.ActivityPrizePoolRefreshLockKey(), token, 30*time.Second)
	if err != nil || !locked {
		return err
	}
	defer s.activityCache.ReleaseLock(cache.ActivityPrizePoolRefreshLockKey(), token)

	now := time.Now()
	activity, err := s.repo.FindActive(now)
	if err != nil {
		return nil
	}
	prizes, err := s.repo.ListActivePrizes(activity.ID)
	if err != nil {
		return err
	}
	poolKey := cache.ActivityPrizePoolKey(activity.ID)
	for _, prize := range prizes {
		if prize.TotalNum < 0 {
			continue
		}
		field := strconv.FormatUint(prize.ID, 10)
		planAmount := parseDailyRelease(prize.ReleasePlan)
		if _, ok, err := s.activityCache.HGetInt(poolKey, field); err != nil {
			return err
		} else if !ok {
			rebuildAmount := prize.LeftNum
			if planAmount > 0 {
				rebuildAmount = 0
			}
			if err := s.activityCache.HSetInt(poolKey, field, rebuildAmount); err != nil {
				return err
			}
		}
		if planAmount <= 0 {
			continue
		}
		windowKey := now.Format("20060102")
		if _, err := s.repo.FindReleaseState(prize.ID, windowKey); err == nil {
			continue
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		add := minInt64(planAmount, prize.LeftNum)
		if add <= 0 {
			continue
		}
		err = s.db.Transaction(func(tx *gorm.DB) error {
			if err := s.repo.UpsertReleaseState(tx, prize.ID, windowKey, add); err != nil {
				return err
			}
			_, err := s.activityCache.HIncrBy(poolKey, field, add)
			return err
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ActivityLotteryService) activePrizes(activityID uint64) ([]model.ActivityLotteryPrize, error) {
	key := cache.ActivityConfigCacheKey(activityID)
	var prizes []model.ActivityLotteryPrize
	if ok, err := s.activityCache.GetJSON(key, &prizes); err != nil {
		return nil, err
	} else if ok {
		return prizes, nil
	}
	prizes, err := s.repo.ListActivePrizes(activityID)
	if err != nil {
		return nil, err
	}
	_ = s.activityCache.SetJSON(key, prizes, 5*time.Minute)
	return prizes, nil
}

func (s *ActivityLotteryService) quotaAndEligibility(playerID uint64, ip string, activity *model.ActivityLottery, now time.Time) (*ActivityQuotaView, bool, string, error) {
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dayEnd := dayStart.AddDate(0, 0, 1)
	dayKey := dayStart.Format("20060102")

	playerUsed, _, err := s.activityCache.GetInt(cache.ActivityDailyPlayerCounterKey(activity.ID, playerID, dayKey))
	if err != nil {
		return nil, false, "", err
	}
	if playerUsed == 0 {
		playerUsed, err = s.repo.CountPlayerRecordsOnDay(playerID, activity.ID, dayStart, dayEnd)
		if err != nil {
			return nil, false, "", err
		}
		_ = s.activityCache.SetInt(cache.ActivityDailyPlayerCounterKey(activity.ID, playerID, dayKey), playerUsed, time.Until(dayEnd))
	}

	ipUsed, _, err := s.activityCache.GetInt(cache.ActivityDailyIPCounterKey(activity.ID, ip, dayKey))
	if err != nil {
		return nil, false, "", err
	}
	if ipUsed == 0 && ip != "" {
		ipUsed, err = s.repo.CountIPRecordsOnDay(ip, activity.ID, dayStart, dayEnd)
		if err != nil {
			return nil, false, "", err
		}
		_ = s.activityCache.SetInt(cache.ActivityDailyIPCounterKey(activity.ID, ip, dayKey), ipUsed, time.Until(dayEnd))
	}

	quota := &ActivityQuotaView{
		DailyLimit:     activity.DailyLimit,
		UsedToday:      uint32(playerUsed),
		RemainingToday: remaining(activity.DailyLimit, uint32(playerUsed)),
		IPDailyLimit:   activity.IPDailyLimit,
		IPUsedToday:    uint32(ipUsed),
	}
	if uint32(playerUsed) >= activity.DailyLimit {
		return quota, false, "daily draw limit reached", nil
	}
	if activity.IPDailyLimit > 0 && uint32(ipUsed) >= activity.IPDailyLimit {
		return quota, false, "ip draw limit reached", nil
	}
	if ok, err := s.blacklisted("player", strconv.FormatUint(playerID, 10), now); err != nil || ok {
		return quota, false, "player blacklisted", err
	}
	if ip != "" {
		if ok, err := s.blacklisted("ip", ip, now); err != nil || ok {
			return quota, false, "ip blacklisted", err
		}
	}
	return quota, true, "", nil
}

func (s *ActivityLotteryService) blacklisted(targetType string, target string, now time.Time) (bool, error) {
	key := cache.ActivityBlacklistKey(targetType, target)
	if value, ok, err := s.activityCache.GetInt(key); err != nil {
		return false, err
	} else if ok {
		return value == 1, nil
	}
	_, err := s.repo.FindActiveBlacklist(targetType, target, now)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			_ = s.activityCache.SetInt(key, 0, time.Minute)
			return false, nil
		}
		return false, err
	}
	_ = s.activityCache.SetInt(key, 1, time.Minute)
	return true, nil
}

func (s *ActivityLotteryService) pickPrize(activityID uint64, prizes []model.ActivityLotteryPrize) (model.ActivityLotteryPrize, uint32, error) {
	weights := make([]uint32, 0, len(prizes))
	candidates := make([]model.ActivityLotteryPrize, 0, len(prizes))
	for _, prize := range prizes {
		if prize.Weight == 0 {
			continue
		}
		candidates = append(candidates, prize)
		weights = append(weights, prize.Weight)
	}
	idx := random.WeightIndex(weights)
	if idx < 0 {
		return model.ActivityLotteryPrize{}, 0, errors.New("activity prize weight invalid")
	}
	total := uint32(0)
	for _, weight := range weights {
		total += weight
	}
	point := uint32(rand.Intn(int(total))) + 1
	return candidates[idx], point, nil
}

func (s *ActivityLotteryService) reserveRedisPrize(activityID uint64, prizeID uint64) (bool, error) {
	if !s.activityCache.Available() {
		return true, nil
	}
	count, err := s.activityCache.HIncrBy(cache.ActivityPrizePoolKey(activityID), strconv.FormatUint(prizeID, 10), -1)
	if err != nil {
		return false, err
	}
	if count < 0 {
		_, _ = s.activityCache.HIncrBy(cache.ActivityPrizePoolKey(activityID), strconv.FormatUint(prizeID, 10), 1)
		return false, nil
	}
	return true, nil
}

func (s *ActivityLotteryService) fallbackPrize(prizes []model.ActivityLotteryPrize) (model.ActivityLotteryPrize, error) {
	for _, prize := range prizes {
		if prize.Fallback == 1 {
			return prize, nil
		}
	}
	for _, prize := range prizes {
		if prize.TotalNum < 0 {
			return prize, nil
		}
	}
	return model.ActivityLotteryPrize{}, errors.New("activity fallback prize missing")
}

func (s *ActivityLotteryService) markMessageFailed(businessID string, reason string) {
	_ = s.db.Transaction(func(tx *gorm.DB) error {
		message, err := s.repo.FindLocalMessageForUpdate(tx, businessID)
		if err != nil {
			return err
		}
		message.Status = MessageFailed
		message.RetryCount++
		message.LastError = reason
		next := time.Now().Add(time.Duration(message.RetryCount+1) * time.Minute)
		message.NextRetryAt = &next
		return s.repo.SaveLocalMessage(tx, message)
	})
}

func (s *ActivityLotteryService) bumpCounters(activity *model.ActivityLottery, playerID uint64, ip string) {
	now := time.Now()
	dayEnd := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)
	dayKey := now.Format("20060102")
	ttl := time.Until(dayEnd)
	_, _ = s.activityCache.IncrWithTTL(cache.ActivityDailyPlayerCounterKey(activity.ID, playerID, dayKey), ttl)
	if ip != "" {
		_, _ = s.activityCache.IncrWithTTL(cache.ActivityDailyIPCounterKey(activity.ID, ip, dayKey), ttl)
	}
}

func (s *ActivityLotteryService) activityView(activity *model.ActivityLottery) *ActivityView {
	return &ActivityView{
		ID:          activity.ID,
		Code:        activity.Code,
		Name:        activity.Name,
		Description: activity.Description,
		BannerImage: activity.BannerImage,
		StartAt:     activity.StartAt.Format(time.RFC3339),
		EndAt:       activity.EndAt.Format(time.RFC3339),
	}
}

func (s *ActivityLotteryService) prizeViews(activityID uint64, prizes []model.ActivityLotteryPrize) []ActivityPrizeView {
	views := make([]ActivityPrizeView, 0, len(prizes))
	for _, prize := range prizes {
		views = append(views, s.prizeView(activityID, prize))
	}
	return views
}

func (s *ActivityLotteryService) prizeView(activityID uint64, prize model.ActivityLotteryPrize) ActivityPrizeView {
	left := prize.LeftNum
	if prize.TotalNum >= 0 {
		if value, ok, err := s.activityCache.HGetInt(cache.ActivityPrizePoolKey(activityID), strconv.FormatUint(prize.ID, 10)); err == nil && ok {
			left = value
		}
	}
	return ActivityPrizeView{
		ID:           prize.ID,
		Name:         prize.Name,
		Description:  prize.Description,
		Icon:         prize.Icon,
		RewardType:   prize.RewardType,
		RewardID:     prize.RewardID,
		RewardCount:  prize.RewardCount,
		Quality:      prize.Quality,
		LeftNum:      left,
		Unlimited:    prize.TotalNum < 0,
		DisplayOrder: prize.DisplayOrder,
	}
}

func (s *ActivityLotteryService) recordViews(records []model.ActivityLotteryRecord) []ActivityRecordView {
	views := make([]ActivityRecordView, 0, len(records))
	for _, record := range records {
		views = append(views, ActivityRecordView{
			DrawNo:         record.DrawNo,
			PrizeName:      record.PrizeName,
			RewardType:     record.RewardType,
			RewardID:       record.RewardID,
			RewardCount:    record.RewardCount,
			DeliveryStatus: record.DeliveryStatus,
			CreatedAt:      record.CreatedAt.Format(time.RFC3339),
		})
	}
	return views
}

func remaining(limit uint32, used uint32) uint32 {
	if used >= limit {
		return 0
	}
	return limit - used
}

func parseDailyRelease(plan string) int64 {
	if len(plan) <= len("daily:") || plan[:len("daily:")] != "daily:" {
		return 0
	}
	value, err := strconv.ParseInt(plan[len("daily:"):], 10, 64)
	if err != nil {
		return 0
	}
	return value
}

func minInt64(a int64, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
