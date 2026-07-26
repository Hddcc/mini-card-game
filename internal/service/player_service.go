package service

import (
	"errors"
	"strings"
	"time"
	"unicode"

	"mini-card-game/internal/model"
	"mini-card-game/internal/repository"

	"gorm.io/gorm"
)

const (
	MaxStamina                    uint32 = 120
	StaminaRecoverSeconds                = 5 * 60
	MaxPlayerNameDailyChanges     uint32 = 3
	PlayerNameEmptyMessage               = "用户名不能为空"
	PlayerNameLengthMessage              = "用户名长度需为 2 - 16 个字符"
	PlayerNameSpacingMessage             = "用户名首尾不能包含空格"
	PlayerNameInvalidCharsMessage        = "用户名仅支持中文、英文、数字"
	PlayerNameSensitiveMessage           = "用户名包含违规内容，请重新输入"
	PlayerNameDailyLimitMessage          = "今日修改次数已达上限，请明日再试"
)

var (
	ErrPlayerNameEmpty        = errors.New(PlayerNameEmptyMessage)
	ErrPlayerNameLength       = errors.New(PlayerNameLengthMessage)
	ErrPlayerNameSpacing      = errors.New(PlayerNameSpacingMessage)
	ErrPlayerNameInvalidChars = errors.New(PlayerNameInvalidCharsMessage)
	ErrPlayerNameSensitive    = errors.New(PlayerNameSensitiveMessage)
	ErrPlayerNameDailyLimit   = errors.New(PlayerNameDailyLimitMessage)
)

type PlayerService struct {
	playerRepo *repository.PlayerRepository
	db         *gorm.DB
	now        func() time.Time
}

func NewPlayerService(db *gorm.DB, playerRepo *repository.PlayerRepository) *PlayerService {
	return &PlayerService{playerRepo: playerRepo, db: db, now: time.Now}
}

type ProfileView struct {
	UniqueID  uint64    `json:"unique_id"`
	PlayerID  uint64    `json:"player_id"`
	UserID    uint64    `json:"user_id"`
	Name      string    `json:"name"`
	Nickname  string    `json:"nickname"`
	Level     uint32    `json:"level"`
	Exp       uint32    `json:"exp"`
	Avatar    string    `json:"avatar"`
	Power     uint64    `json:"power"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	LegacyID        uint64    `json:"ID"`
	LegacyUserID    uint64    `json:"UserID"`
	LegacyNickname  string    `json:"Nickname"`
	LegacyLevel     uint32    `json:"Level"`
	LegacyExp       uint32    `json:"Exp"`
	LegacyAvatar    string    `json:"Avatar"`
	LegacyPower     uint64    `json:"Power"`
	LegacyCreatedAt time.Time `json:"CreatedAt"`
	LegacyUpdatedAt time.Time `json:"UpdatedAt"`
}

func BuildProfileView(profile *model.PlayerProfile) *ProfileView {
	return &ProfileView{
		UniqueID:        profile.ID,
		PlayerID:        profile.ID,
		UserID:          profile.UserID,
		Name:            profile.Nickname,
		Nickname:        profile.Nickname,
		Level:           profile.Level,
		Exp:             profile.Exp,
		Avatar:          profile.Avatar,
		Power:           profile.Power,
		CreatedAt:       profile.CreatedAt,
		UpdatedAt:       profile.UpdatedAt,
		LegacyID:        profile.ID,
		LegacyUserID:    profile.UserID,
		LegacyNickname:  profile.Nickname,
		LegacyLevel:     profile.Level,
		LegacyExp:       profile.Exp,
		LegacyAvatar:    profile.Avatar,
		LegacyPower:     profile.Power,
		LegacyCreatedAt: profile.CreatedAt,
		LegacyUpdatedAt: profile.UpdatedAt,
	}
}

func (s *PlayerService) Profile(playerID uint64) (*ProfileView, error) {
	profile, err := s.playerRepo.FindProfile(playerID)
	if err != nil {
		return nil, err
	}
	return BuildProfileView(profile), nil
}

func (s *PlayerService) UpdateName(playerID uint64, name string) (*ProfileView, error) {
	if err := ValidatePlayerName(name); err != nil {
		return nil, err
	}

	now := s.now()
	var updated *model.PlayerProfile
	err := s.db.Transaction(func(tx *gorm.DB) error {
		profile, err := s.playerRepo.LockProfile(tx, playerID)
		if err != nil {
			return err
		}

		changeCount := profile.NameDailyChangeCount
		if profile.NameChangeDate == nil || !sameLocalDay(*profile.NameChangeDate, now) {
			changeCount = 0
		}
		if changeCount >= MaxPlayerNameDailyChanges {
			return ErrPlayerNameDailyLimit
		}

		day := localDayStart(now)
		profile.Nickname = name
		profile.NameUpdatedAt = &now
		profile.NameChangeDate = &day
		profile.NameDailyChangeCount = changeCount + 1
		if err := s.playerRepo.UpdateProfileName(tx, profile); err != nil {
			return err
		}
		updated = profile
		return nil
	})
	if err != nil {
		return nil, err
	}
	return BuildProfileView(updated), nil
}

func ValidatePlayerName(name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrPlayerNameEmpty
	}
	if strings.TrimSpace(name) != name {
		return ErrPlayerNameSpacing
	}
	runes := []rune(name)
	if len(runes) < 2 || len(runes) > 16 {
		return ErrPlayerNameLength
	}
	for _, r := range runes {
		if !isAllowedPlayerNameRune(r) {
			return ErrPlayerNameInvalidChars
		}
	}
	if containsSensitivePlayerNameTerm(name) {
		return ErrPlayerNameSensitive
	}
	return nil
}

func IsPlayerNameUpdateClientError(err error) bool {
	return errors.Is(err, ErrPlayerNameEmpty) ||
		errors.Is(err, ErrPlayerNameLength) ||
		errors.Is(err, ErrPlayerNameSpacing) ||
		errors.Is(err, ErrPlayerNameInvalidChars) ||
		errors.Is(err, ErrPlayerNameSensitive) ||
		errors.Is(err, ErrPlayerNameDailyLimit)
}

func isAllowedPlayerNameRune(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		unicode.In(r, unicode.Han)
}

func containsSensitivePlayerNameTerm(name string) bool {
	normalized := strings.ToLower(name)
	terms := []string{
		"admin",
		"gm",
		"official",
		"system",
		"badword",
		"官方",
		"客服",
		"系统",
		"管理员",
		"敏感",
		"违禁",
		"赌博",
		"色情",
		"政治",
	}
	for _, term := range terms {
		if strings.Contains(normalized, term) {
			return true
		}
	}
	return false
}

func sameLocalDay(a time.Time, b time.Time) bool {
	ay, am, ad := a.In(time.Local).Date()
	by, bm, bd := b.In(time.Local).Date()
	return ay == by && am == bm && ad == bd
}

func localDayStart(t time.Time) time.Time {
	local := t.In(time.Local)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, local.Location())
}

func (s *PlayerService) Assets(playerID uint64) (interface{}, error) {
	asset, err := s.playerRepo.FindAsset(playerID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	changed := SettleStamina(asset, now)
	if changed {
		if err := s.playerRepo.SaveAsset(asset); err != nil {
			return nil, err
		}
	}
	return BuildAssetView(asset, now), nil
}

type AssetView struct {
	PlayerID            uint64 `json:"player_id"`
	Gold                uint64 `json:"gold"`
	Diamond             uint64 `json:"diamond"`
	Stamina             uint32 `json:"stamina"`
	MaxStamina          uint32 `json:"max_stamina"`
	NextStaminaSeconds  int64  `json:"next_stamina_seconds"`
	StaminaRecoverEvery int64  `json:"stamina_recover_every"`
}

func BuildAssetView(asset *model.PlayerAsset, now time.Time) AssetView {
	return AssetView{
		PlayerID:            asset.PlayerID,
		Gold:                asset.Gold,
		Diamond:             asset.Diamond,
		Stamina:             asset.Stamina,
		MaxStamina:          MaxStamina,
		NextStaminaSeconds:  NextStaminaSeconds(asset, now),
		StaminaRecoverEvery: StaminaRecoverSeconds,
	}
}

func SettleStamina(asset *model.PlayerAsset, now time.Time) bool {
	if asset.Stamina >= MaxStamina {
		if asset.Stamina > MaxStamina {
			asset.Stamina = MaxStamina
			return true
		}
		return false
	}
	if asset.StaminaUpdatedAt == nil {
		asset.StaminaUpdatedAt = &now
		return true
	}
	elapsed := int64(now.Sub(*asset.StaminaUpdatedAt).Seconds())
	if elapsed < StaminaRecoverSeconds {
		return false
	}
	recovered := uint32(elapsed / StaminaRecoverSeconds)
	if recovered == 0 {
		return false
	}
	asset.Stamina += recovered
	if asset.Stamina >= MaxStamina {
		asset.Stamina = MaxStamina
		asset.StaminaUpdatedAt = nil
		return true
	}
	next := asset.StaminaUpdatedAt.Add(time.Duration(recovered*StaminaRecoverSeconds) * time.Second)
	asset.StaminaUpdatedAt = &next
	return true
}

func TouchStaminaAfterSpend(asset *model.PlayerAsset, now time.Time) {
	if asset.Stamina < MaxStamina && asset.StaminaUpdatedAt == nil {
		asset.StaminaUpdatedAt = &now
	}
	if asset.Stamina >= MaxStamina {
		asset.StaminaUpdatedAt = nil
	}
}

func NextStaminaSeconds(asset *model.PlayerAsset, now time.Time) int64 {
	if asset.Stamina >= MaxStamina || asset.StaminaUpdatedAt == nil {
		return 0
	}
	nextAt := asset.StaminaUpdatedAt.Add(StaminaRecoverSeconds * time.Second)
	seconds := int64(nextAt.Sub(now).Seconds())
	if seconds < 0 {
		return 0
	}
	return seconds
}
