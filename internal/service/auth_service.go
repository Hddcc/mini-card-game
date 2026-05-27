package service

import (
	"errors"
	"mini-card-game/internal/config"
	"mini-card-game/internal/model"
	"mini-card-game/internal/pkg/jwt"
	"mini-card-game/internal/pkg/password"
	"mini-card-game/internal/repository"

	"gorm.io/gorm"
)

type AuthService struct {
	cfg        *config.Config
	db         *gorm.DB
	userRepo   *repository.UserRepository
	playerRepo *repository.PlayerRepository
}

func NewAuthService(cfg *config.Config, db *gorm.DB, userRepo *repository.UserRepository, playerRepo *repository.PlayerRepository) *AuthService {
	return &AuthService{
		cfg:        cfg,
		db:         db,
		userRepo:   userRepo,
		playerRepo: playerRepo,
	}
}

type RegisterInput struct {
	Username string
	Password string
	Nickname string
}

func (s *AuthService) Register(input RegisterInput) (uint64, uint64, error) {
	if _, err := s.userRepo.FindByUsername(input.Username); err == nil {
		return 0, 0, errors.New("username already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, 0, err
	}

	hash, err := password.Hash(input.Password)
	if err != nil {
		return 0, 0, err
	}

	var userID uint64
	var playerID uint64

	err = s.db.Transaction(func(tx *gorm.DB) error {
		user := &model.User{
			Username:     input.Username,
			PasswordHash: hash,
			Status:       1,
		}
		if err := s.userRepo.Create(tx, user); err != nil {
			return err
		}

		profile := &model.PlayerProfile{
			UserID:   user.ID,
			Nickname: input.Nickname,
			Level:    1,
			Exp:      0,
			Avatar:   "avatar_001",
			Power:    0,
		}
		if err := s.playerRepo.CreateProfile(tx, profile); err != nil {
			return err
		}

		asset := &model.PlayerAsset{
			PlayerID: profile.ID,
			Gold:     10000,
			Diamond:  2800,
			Stamina:  120,
		}
		if err := s.playerRepo.CreateAsset(tx, asset); err != nil {
			return err
		}

		userID = user.ID
		playerID = profile.ID
		return nil
	})
	if err != nil {
		return 0, 0, err
	}

	return userID, playerID, nil
}

type LoginInput struct {
	Username string
	Password string
}

type LoginOutput struct {
	Token     string
	ExpiresIn int64
	PlayerID  uint64
	Nickname  string
	Level     uint32
}

func (s *AuthService) Login(input LoginInput) (*LoginOutput, error) {
	user, err := s.userRepo.FindByUsername(input.Username)
	if err != nil {
		return nil, errors.New("username or password invalid")
	}

	if !password.Check(user.PasswordHash, input.Password) {
		return nil, errors.New("username or password invalid")
	}

	profile, err := s.playerRepo.FindProfileByUserID(user.ID)
	if err != nil {
		return nil, err
	}

	token, err := jwt.Generate(s.cfg.JWTSecret, s.cfg.JWTExpireSeconds, user.ID, profile.ID)
	if err != nil {
		return nil, err
	}
	return &LoginOutput{
		Token:     token,
		ExpiresIn: s.cfg.JWTExpireSeconds,
		PlayerID:  profile.ID,
		Nickname:  profile.Nickname,
		Level:     profile.Level,
	}, nil
}
