package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"blog-api/internal/model"
	"blog-api/internal/repository"
	"blog-api/pkg/auth"
	"blog-api/pkg/exception"

	"github.com/gofrs/uuid/v5"
)

type UserService struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	jwtManager       *auth.JWTManager
	passwordManager  *auth.PasswordManager
}

func NewUserService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	jwtManager *auth.JWTManager,
	passwordManager *auth.PasswordManager,
) *UserService {
	return &UserService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtManager:       jwtManager,
		passwordManager:  passwordManager,
	}
}

func (s *UserService) Register(
	ctx context.Context,
	req *model.UserCreateRequest,
) (*model.TokenResponse, *exception.ApiError) {
	exists, err := s.userRepo.ExistsByField(ctx, "email", req.Email)
	if err != nil {
		log.Printf("failed to check user existance: %v", err)
		return nil, exception.DatabaseError(err.Error())
	}
	if exists {
		return nil, exception.ConflictError("email already exists")
	}
	exists, err = s.userRepo.ExistsByField(ctx, "username", req.Username)
	if err != nil {
		log.Printf("failed to check user existance: %v", err)
		return nil, exception.DatabaseError(err.Error())
	}
	if exists {
		return nil, exception.ConflictError("username already exists")
	}

	if err := s.passwordManager.ValidatePasswordStrength(req.Password); err != nil {
		return nil, exception.BadRequestError(err.Error())
	}
	hashedPassword, err := s.passwordManager.HashPassword(req.Password)
	if err != nil {
		log.Printf("failed to hash a password: %v", err)
		return nil, exception.InternalServerError("failed to hash password")
	}

	user := &model.User{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: hashedPassword,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		log.Printf("failed to create a user: %v", err)
		return nil, exception.DatabaseError(err.Error())
	}

	return s.createTokenPair(ctx, user.ID)
}

func (s *UserService) Login(
	ctx context.Context,
	req *model.UserLoginRequest,
) (*model.TokenResponse, *exception.ApiError) {
	user, err := s.userRepo.GetByField(ctx, "email", req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			// Do not reveal if email exists or not
			return nil, exception.BadRequestError("invalid email or password")
		}
		log.Printf("failed to fetch a user: %v", err)
		return nil, exception.DatabaseError(err.Error())
	}

	if !s.passwordManager.CheckPassword(req.Password, user.PasswordHash) {
		return nil, exception.BadRequestError("invalid email or password")
	}

	return s.createTokenPair(ctx, user.ID)
}

func (s *UserService) RefreshToken(
	ctx context.Context,
	req *model.RefreshTokenRequest,
) (*model.TokenResponse, *exception.ApiError) {
	tokenUUID, err := uuid.FromString(req.RefreshToken)
	if err != nil {
		return nil, exception.BadRequestError("failed to parse the refresh token value as UUID")
	}

	rt, err := s.refreshTokenRepo.GetByValue(ctx, tokenUUID)
	if err != nil {
		if errors.Is(err, repository.ErrRefreshTokenNotFound) {
			return nil, exception.BadRequestError("refresh token not found")
		}
		log.Printf("failed to fetch a refresh token: %v", err)
		return nil, exception.DatabaseError(err.Error())
	}

	if time.Now().After(rt.ExpiresAt) {
		if err := s.refreshTokenRepo.DeleteByValue(ctx, tokenUUID); err != nil {
			log.Printf("failed to delete expired refresh token: %v", err)
		}
		return nil, exception.BadRequestError("refresh token expired")
	}

	if err := s.refreshTokenRepo.DeleteByValue(ctx, tokenUUID); err != nil {
		log.Printf("failed to delete expired refresh token: %v", err)
		return nil, exception.DatabaseError(err.Error())
	}

	return s.createTokenPair(ctx, rt.UserID)
}

func (s *UserService) createTokenPair(ctx context.Context, userID int) (*model.TokenResponse, *exception.ApiError) {
	accessToken, accessExpiresAt, err := s.jwtManager.GenerateToken(userID)
	if err != nil {
		return nil, exception.InternalServerError("failed to generate access token")
	}

	refreshToken := &model.RefreshToken{
		Value:     uuid.Must(uuid.NewV4()),
		UserID:    userID,
		ExpiresAt: time.Now().Add(s.jwtManager.RefreshTokenTTL),
	}
	if err := s.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		return nil, exception.DatabaseError(err.Error())
	}

	return &model.TokenResponse{
		AccessToken:        accessToken,
		AccessTokenExpiry:  accessExpiresAt,
		RefreshToken:       refreshToken.Value.String(),
		RefreshTokenExpiry: refreshToken.ExpiresAt,
	}, nil
}

func (s *UserService) GetByID(ctx context.Context, id int) (*model.User, *exception.ApiError) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, exception.NotFoundError(fmt.Sprintf("user with id=%d not found", id))
		}
		return nil, exception.DatabaseError(fmt.Sprintf("failed to get user by id=%d: %v", id, err))
	}
	return user, nil
}
