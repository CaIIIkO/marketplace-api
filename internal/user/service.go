package user

import (
	"context"
	"errors"
	"marketplace-api/internal/auth"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

var (
	loginRegex    = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_.]{1,28}[a-zA-Z0-9]$`)
	passwordRegex = regexp.MustCompile(`^[a-zA-Z0-9!@#$%^&*()_+\-=\[\]{};':"\\|,.<>/?]{6,30}$`)
)

type RepositoryInterface interface {
	Create(ctx context.Context, u *User) (*User, error)
	GetByLogin(ctx context.Context, login string) (*User, error)
}

type Service struct {
	repo       RepositoryInterface
	jwtManager *auth.JWTManager
}

func NewUserService(repo RepositoryInterface, jwtManager *auth.JWTManager) *Service {
	return &Service{repo: repo, jwtManager: jwtManager}
}

// Register - регистрация пользователя
func (s *Service) Register(ctx context.Context, input *RegisterRequest) (*User, error) {
	if err := s.validateRegisterInput(ctx, input); err != nil {
		return nil, err
	}

	//Хэширование пароля
	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		Login:        input.Login,
		PasswordHash: string(hashed),
	}

	user, err = s.repo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// validateRegisterInput проверяет корректность входных данных при регистрации
func (s *Service) validateRegisterInput(ctx context.Context, input *RegisterRequest) error {
	// Валидация логина
	if !loginRegex.MatchString(input.Login) {
		return errors.New("invalid login: must be 3-30 characters (letters, numbers, underscore, dots)")
	}

	if strings.Contains(input.Login, "..") || strings.Contains(input.Login, "__") ||
		strings.Contains(input.Login, "_.") || strings.Contains(input.Login, "._") {
		return errors.New("invalid login: must not contain repeated underscores and dots")
	}

	//Проверка уникальности login
	existing, err := s.repo.GetByLogin(ctx, input.Login)
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.New("user login already exists")
	}

	// Валидация пароля
	if len(input.Password) < 6 || len(input.Password) > 30 {
		return errors.New("invalid password: must be at least 6 - 30 characters")
	}

	if !passwordRegex.MatchString(input.Password) {
		return errors.New(`invalid password: ust contain only a-zA-Z0-9!@#$%^&*()_+\-=\[\]{};':"\\|,.<>/?`)
	}

	var hasUpper, hasLower, hasDigit bool

	for _, ch := range input.Password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		}
	}

	if !hasUpper {
		return errors.New("invalid password: must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("invalid password: must contain at least one lowercase letter")
	}
	if !hasDigit {
		return errors.New("invalid password: must contain at least one number")
	}

	return nil
}

// Authenticate - аутентификация пользователя
func (s *Service) Authenticate(ctx context.Context, input *LoginRequest) (string, error) {
	user, err := s.validateAuthenticateInput(ctx, input)
	if err != nil {
		return "", err
	}

	token, err := s.jwtManager.Generate(user.ID)
	if err != nil {
		return "", errors.New("token error")
	}

	return token, nil
}

// validateLoginInput проверяет корректность входных данных при аутентификации
func (s *Service) validateAuthenticateInput(ctx context.Context, input *LoginRequest) (*User, error) {
	//Проверка существования пользователя
	user, err := s.repo.GetByLogin(ctx, input.Login)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	//Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, errors.New("invalid password")
	}
	return user, err
}
