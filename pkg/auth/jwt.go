package auth

import (
	"blog-api/pkg/settings"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// types
type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

// errors
var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

// config
type JWTManagerConfig struct {
	JWTSecret             string
	AccessTokenTTLMunutes int
	RefreshTokenTTLHours  int
}

func (c *JWTManagerConfig) Setup() []settings.EnvLoadable {
	return []settings.EnvLoadable{
		settings.Item[string]{Name: "JWT_SECRET", Default: settings.NoDefault, Field: &c.JWTSecret},
		settings.Item[int]{Name: "JWT_ACCESS_TOKEN_TTL_MINUTES", Default: 5, Field: &c.AccessTokenTTLMunutes},
		settings.Item[int]{Name: "JWT_REFRESH_TOKEN_TTL_HOURS", Default: 24, Field: &c.RefreshTokenTTLHours},
	}
}

// manager
type JWTManager struct {
	secretKey []byte
	ttl       time.Duration
}

func NewJWTManager(secretKey string, tokenTTLMinutes int) *JWTManager {
	return &JWTManager{
		[]byte(secretKey),
		time.Duration(tokenTTLMinutes) * time.Minute,
	}
}

func (m *JWTManager) GenerateToken(userID int) (string, time.Time, error) {
	var expires time.Time = time.Now().Add(m.ttl * time.Minute)
	claims := Claims{
		UserID: userID,
	}
	claims.IssuedAt = jwt.NewNumericDate(time.Now())
	claims.ExpiresAt = jwt.NewNumericDate(expires)
	var token *jwt.Token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign a token: %w", err)
	}
	return signed, expires, nil
}

func (m *JWTManager) KeyFunc(token *jwt.Token) (any, error) {
	if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
		return nil, fmt.Errorf("wrong encryption algorythm: %v", token.Header["alg"])
	}
	return m.secretKey, nil
}

func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	claims := Claims{}
	if _, err := jwt.ParseWithClaims(tokenString, &claims, m.KeyFunc); err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return &claims, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}
	return &claims, nil
}

func (m *JWTManager) RefreshToken(tokenString string) (string, time.Time, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil && !errors.Is(err, ErrExpiredToken) {
		return "", time.Time{}, err
	}
	return m.GenerateToken(claims.UserID)
}

// unsafe user_id lookup (never use in auth)
func (m *JWTManager) GetUserIDFromToken(tokenString string) (int, error) {
	token, _, err := (&jwt.Parser{}).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return 0, err
	}
	claims := token.Claims.(jwt.MapClaims)
	userID, ok := claims["user_id"].(float64)
	if !ok {
		return 0, fmt.Errorf("user_id not found or invalid type")
	}
	return int(userID), nil
}
