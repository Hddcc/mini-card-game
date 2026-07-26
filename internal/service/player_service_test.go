package service

import (
	"errors"
	"testing"
	"time"

	"mini-card-game/internal/model"
	"mini-card-game/internal/repository"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type playerServiceTestEnv struct {
	db  *gorm.DB
	svc *PlayerService
}

func newPlayerServiceTestEnv(t *testing.T) *playerServiceTestEnv {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.PlayerProfile{}); err != nil {
		t.Fatalf("migrate profile: %v", err)
	}
	repo := repository.NewPlayerRepository(db)
	return &playerServiceTestEnv{db: db, svc: NewPlayerService(db, repo)}
}

func seedPlayerProfile(t *testing.T, db *gorm.DB, userID uint64, name string) model.PlayerProfile {
	t.Helper()
	profile := model.PlayerProfile{
		UserID:   userID,
		Nickname: name,
		Level:    2,
		Exp:      120,
		Avatar:   "avatar_001",
		Power:    360,
	}
	if err := db.Create(&profile).Error; err != nil {
		t.Fatalf("seed profile: %v", err)
	}
	return profile
}

func TestPlayerProfileViewFields(t *testing.T) {
	env := newPlayerServiceTestEnv(t)
	profile := seedPlayerProfile(t, env.db, 101, "取经人")

	view, err := env.svc.Profile(profile.ID)
	if err != nil {
		t.Fatalf("profile: %v", err)
	}
	if view.UniqueID != profile.ID || view.PlayerID != profile.ID || view.UserID != profile.UserID {
		t.Fatalf("unexpected identity fields: %#v", view)
	}
	if view.Name != "取经人" || view.Nickname != "取经人" {
		t.Fatalf("unexpected display name fields: %#v", view)
	}
	if view.LegacyID != profile.ID || view.LegacyUserID != profile.UserID || view.LegacyNickname != "取经人" {
		t.Fatalf("legacy fields not preserved: %#v", view)
	}
	if view.Level != 2 || view.Exp != 120 || view.Avatar != "avatar_001" || view.Power != 360 {
		t.Fatalf("profile stats changed: %#v", view)
	}
}

func TestPlayerUpdateNameSuccessPersistsAndAllowsDuplicates(t *testing.T) {
	env := newPlayerServiceTestEnv(t)
	now := time.Date(2026, 6, 28, 10, 0, 0, 0, time.Local)
	env.svc.now = func() time.Time { return now }
	first := seedPlayerProfile(t, env.db, 101, "取经人")
	second := seedPlayerProfile(t, env.db, 102, "巡山人")

	view, err := env.svc.UpdateName(first.ID, "齐天大圣")
	if err != nil {
		t.Fatalf("update name: %v", err)
	}
	if view.UniqueID != first.ID || view.Name != "齐天大圣" || view.Nickname != "齐天大圣" {
		t.Fatalf("unexpected update view: %#v", view)
	}

	dupeView, err := env.svc.UpdateName(second.ID, "齐天大圣")
	if err != nil {
		t.Fatalf("duplicate name should be allowed: %v", err)
	}
	if dupeView.Name != "齐天大圣" {
		t.Fatalf("duplicate update name = %s", dupeView.Name)
	}

	profileView, err := env.svc.Profile(first.ID)
	if err != nil {
		t.Fatalf("reload profile: %v", err)
	}
	if profileView.UniqueID != first.ID || profileView.Name != "齐天大圣" {
		t.Fatalf("persisted profile mismatch: %#v", profileView)
	}

	var stored model.PlayerProfile
	if err := env.db.First(&stored, first.ID).Error; err != nil {
		t.Fatalf("load stored profile: %v", err)
	}
	if stored.NameUpdatedAt == nil || stored.NameChangeDate == nil || stored.NameDailyChangeCount != 1 {
		t.Fatalf("edit metadata not persisted: %#v", stored)
	}
}

func TestValidatePlayerName(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{name: "", err: ErrPlayerNameEmpty},
		{name: "   ", err: ErrPlayerNameEmpty},
		{name: "悟", err: ErrPlayerNameLength},
		{name: "一二三四五六七八九十一二三四五六七", err: ErrPlayerNameLength},
		{name: " 悟空", err: ErrPlayerNameSpacing},
		{name: "悟空 ", err: ErrPlayerNameSpacing},
		{name: "悟空!", err: ErrPlayerNameInvalidChars},
		{name: "系统用户", err: ErrPlayerNameSensitive},
		{name: "admin!", err: ErrPlayerNameInvalidChars},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePlayerName(tt.name)
			if !errors.Is(err, tt.err) {
				t.Fatalf("expected %v, got %v", tt.err, err)
			}
		})
	}
	if err := ValidatePlayerName("Wukong88"); err != nil {
		t.Fatalf("english and digits should be valid: %v", err)
	}
	if err := ValidatePlayerName("悟空88"); err != nil {
		t.Fatalf("mixed chinese and digits should be valid: %v", err)
	}
}

func TestPlayerUpdateNameDailyLimitAndNextDayReset(t *testing.T) {
	env := newPlayerServiceTestEnv(t)
	now := time.Date(2026, 6, 28, 10, 0, 0, 0, time.Local)
	env.svc.now = func() time.Time { return now }
	profile := seedPlayerProfile(t, env.db, 101, "取经人")

	for _, name := range []string{"悟空一号", "悟空二号", "悟空三号"} {
		if _, err := env.svc.UpdateName(profile.ID, name); err != nil {
			t.Fatalf("daily update %s: %v", name, err)
		}
	}
	_, err := env.svc.UpdateName(profile.ID, "悟空四号")
	if !errors.Is(err, ErrPlayerNameDailyLimit) {
		t.Fatalf("expected daily limit, got %v", err)
	}

	reloaded, err := env.svc.Profile(profile.ID)
	if err != nil {
		t.Fatalf("profile after limit: %v", err)
	}
	if reloaded.Name != "悟空三号" {
		t.Fatalf("limited update changed stored name: %s", reloaded.Name)
	}

	now = now.Add(24 * time.Hour)
	view, err := env.svc.UpdateName(profile.ID, "悟空新日")
	if err != nil {
		t.Fatalf("next-day update: %v", err)
	}
	if view.Name != "悟空新日" {
		t.Fatalf("next-day name = %s", view.Name)
	}

	var stored model.PlayerProfile
	if err := env.db.First(&stored, profile.ID).Error; err != nil {
		t.Fatalf("load stored next day: %v", err)
	}
	if stored.NameDailyChangeCount != 1 {
		t.Fatalf("next day count = %d", stored.NameDailyChangeCount)
	}
}
