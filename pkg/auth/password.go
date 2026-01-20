package auth

import (
	"errors"
	"unicode"

	"github.com/tailscale/golang-x-crypto/bcrypt"

	"blog-api/pkg/settings"
)

// errors
var (
	ErrPasswordTooShort = errors.New("password is too short")
	ErrPasswordTooWeak  = errors.New("password too weak")
	ErrPasswordTooLong  = bcrypt.ErrPasswordTooLong
)

// config
type PasswordManagerConfig struct {
	MinLength         int
	Cost              int
	CaseShiftRequired bool
	DigitsRequired    bool
	SymbolsRequired   bool
}

func (c *PasswordManagerConfig) Setup() []settings.EnvLoadable {
	return []settings.EnvLoadable{
		settings.Item[int]{Name: "PASSWORD_MIN_LENGTH", Default: 6, Field: &c.MinLength},
		settings.Item[int]{Name: "PASSWORD_COST", Default: bcrypt.DefaultCost, Field: &c.MinLength},
		settings.Item[bool]{Name: "PASSWORD_MUST_SHIFT_CASE", Default: true, Field: &c.CaseShiftRequired},
		settings.Item[bool]{Name: "PASSWORD_MUST_HAVE_DIGITS", Default: true, Field: &c.DigitsRequired},
		settings.Item[bool]{Name: "PASSWORD_MUST_HAVE_SYMBOLS", Default: true, Field: &c.DigitsRequired},
	}
}

// manager
type PasswordManager struct {
	config *PasswordManagerConfig
}

func NewPasswordManager(config *PasswordManagerConfig) *PasswordManager {
	if config == nil {
		panic("PasswordManager requires a non-nil config")
	}
	return &PasswordManager{config: config}
}

func (pm *PasswordManager) ValidatePasswordStrength(password string) error {
	if len(password) < pm.config.MinLength {
		return ErrPasswordTooShort
	}
	var hasUpper, hasLower, hasDigit, hasSymbol bool
	for _, r := range password {
		if !hasUpper && unicode.IsUpper(r) {
			hasUpper = true
		}
		if !hasLower && unicode.IsLower(r) {
			hasLower = true
		}
		if !hasDigit && unicode.IsDigit(r) {
			hasDigit = true
		}
		if !hasSymbol && (unicode.IsSymbol(r) || unicode.IsPunct(r)) {
			hasSymbol = true
		}
		if (!pm.config.CaseShiftRequired || (hasUpper && hasLower)) &&
			(!pm.config.DigitsRequired || hasDigit) &&
			(!pm.config.SymbolsRequired || hasSymbol) {
			break
		}
	}
	if (pm.config.CaseShiftRequired && (!hasUpper || !hasLower)) ||
		(pm.config.DigitsRequired && !hasDigit) ||
		(pm.config.SymbolsRequired && !hasSymbol) {
		return ErrPasswordTooWeak
	}

	return nil
}

func (pm *PasswordManager) HashPassword(password string) (string, error) {
	if err := pm.ValidatePasswordStrength(password); err != nil {
		return "", err
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), pm.config.Cost)
	if err != nil {
		return "", err
	}
	return string(passwordHash), nil
}

func (pm *PasswordManager) CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
