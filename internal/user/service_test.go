package user_test

import (
	"context"
	"marketplace-api/internal/auth"
	"marketplace-api/internal/user"
	mockuser "marketplace-api/internal/user/mock"
	"regexp"
	"testing"

	uuid "github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

func setupTest(t *testing.T) (*gomock.Controller, *mockuser.MockRepositoryInterface, *user.Service) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockRepo := mockuser.NewMockRepositoryInterface(ctrl)
	jwtManager := auth.NewJWTManager("secret")
	service := user.NewUserService(mockRepo, jwtManager)
	return ctrl, mockRepo, service
}

func TestService_Register(t *testing.T) {
	validInput := &user.RegisterRequest{
		Login:    "Valid_User_123",
		Password: "Syperpassword123",
	}
	hashedRegex := regexp.MustCompile(`^\$2[aby]\$`)

	t.Run("успешная регистрация", func(t *testing.T) {
		ctrl, mockRepo, service := setupTest(t)
		defer ctrl.Finish()

		mockRepo.EXPECT().GetByLogin(gomock.Any(), validInput.Login).Return(nil, nil)

		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, u *user.User) (*user.User, error) {
			assert.Equal(t, validInput.Login, u.Login)
			assert.Regexp(t, hashedRegex, u.PasswordHash)
			u.ID = uuid.New()
			return u, nil
		})

		createdUser, err := service.Register(context.Background(), validInput)
		assert.NoError(t, err)
		assert.Equal(t, validInput.Login, createdUser.Login)
	})

	t.Run("ошибка: логин занят", func(t *testing.T) {
		ctrl, mockRepo, service := setupTest(t)
		defer ctrl.Finish()

		mockRepo.EXPECT().GetByLogin(gomock.Any(), validInput.Login).Return(&user.User{}, nil)

		_, err := service.Register(context.Background(), validInput)
		assert.ErrorContains(t, err, "login already exists")
	})

	t.Run("ошибка: невалидный логин", func(t *testing.T) {
		ctrl, _, service := setupTest(t)
		defer ctrl.Finish()

		bad := *validInput
		bad.Login = "ab"
		_, err := service.Register(context.Background(), &bad)
		assert.ErrorContains(t, err, "invalid login")

		bad.Login = "_asds"
		_, err = service.Register(context.Background(), &bad)
		assert.ErrorContains(t, err, "invalid login")

		bad.Login = "2133213a"
		_, err = service.Register(context.Background(), &bad)
		assert.ErrorContains(t, err, "invalid login")

		bad.Login = "a312__dasd.s"
		_, err = service.Register(context.Background(), &bad)
		assert.ErrorContains(t, err, "invalid login")

		bad.Login = "a1a."
		_, err = service.Register(context.Background(), &bad)
		assert.ErrorContains(t, err, "invalid login")

		bad.Login = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		_, err = service.Register(context.Background(), &bad)
		assert.ErrorContains(t, err, "invalid login")

		bad.Login = "адолвыадолаывдл"
		_, err = service.Register(context.Background(), &bad)
		assert.ErrorContains(t, err, "invalid login")
	})

	t.Run("ошибка: плохой пароль", func(t *testing.T) {
		ctrl, mockRepo, service := setupTest(t)
		defer ctrl.Finish()

		bad := *validInput
		bad.Password = "123"
		mockRepo.EXPECT().GetByLogin(gomock.Any(), validInput.Login).Return(nil, nil)
		_, err := service.Register(context.Background(), &bad)
		assert.ErrorContains(t, err, "invalid password")

		bad.Password = "aaaaaaaaaaaaaaaaa"
		mockRepo.EXPECT().GetByLogin(gomock.Any(), validInput.Login).Return(nil, nil)
		_, err = service.Register(context.Background(), &bad)
		assert.ErrorContains(t, err, "invalid password")

		bad.Password = "123/..456"
		mockRepo.EXPECT().GetByLogin(gomock.Any(), validInput.Login).Return(nil, nil)
		_, err = service.Register(context.Background(), &bad)
		assert.ErrorContains(t, err, "invalid password")

		bad.Password = "dasKJDsas"
		mockRepo.EXPECT().GetByLogin(gomock.Any(), validInput.Login).Return(nil, nil)
		_, err = service.Register(context.Background(), &bad)
		assert.ErrorContains(t, err, "invalid password")

		bad.Password = "ABNSD123AS"
		mockRepo.EXPECT().GetByLogin(gomock.Any(), validInput.Login).Return(nil, nil)
		_, err = service.Register(context.Background(), &bad)
		assert.ErrorContains(t, err, "invalid password")
	})
}

func TestService_Authenticate(t *testing.T) {
	validUser := &user.User{
		ID:           uuid.New(),
		Login:        "Valid_User_123",
		PasswordHash: "$2a$10$vdixmYeT8uaLkOFjwQj./eYgoALkIgL4vKyEbC40a1awkJ/iex4LG", // hash for "syperpassword"
	}

	t.Run("успешная аутентификация", func(t *testing.T) {
		ctrl, mockRepo, service := setupTest(t)
		defer ctrl.Finish()

		mockRepo.EXPECT().GetByLogin(gomock.Any(), validUser.Login).Return(validUser, nil)

		token, err := service.Authenticate(context.Background(), &user.LoginRequest{
			Login:    validUser.Login,
			Password: "syperpassword",
		})

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("ошибка: неверный пароль", func(t *testing.T) {
		ctrl, mockRepo, service := setupTest(t)
		defer ctrl.Finish()

		mockRepo.EXPECT().GetByLogin(gomock.Any(), validUser.Login).Return(validUser, nil)

		_, err := service.Authenticate(context.Background(), &user.LoginRequest{
			Login:    validUser.Login,
			Password: "wrongpassword",
		})

		assert.ErrorContains(t, err, "invalid password")
	})

	t.Run("ошибка: пользователь не найден", func(t *testing.T) {
		ctrl, mockRepo, service := setupTest(t)
		defer ctrl.Finish()

		mockRepo.EXPECT().GetByLogin(gomock.Any(), validUser.Login).Return(nil, nil)

		_, err := service.Authenticate(context.Background(), &user.LoginRequest{
			Login:    validUser.Login,
			Password: "any",
		})

		assert.ErrorContains(t, err, "user not found")
	})

	t.Run("ошибка: токен не сгенерирован", func(t *testing.T) {
		ctrl, mockRepo := gomock.NewController(t), mockuser.NewMockRepositoryInterface(gomock.NewController(t))
		defer ctrl.Finish()

		// Передаём nil jwtManager
		service := user.NewUserService(mockRepo, nil)

		mockRepo.EXPECT().GetByLogin(gomock.Any(), validUser.Login).Return(validUser, nil)

		_, err := service.Authenticate(context.Background(), &user.LoginRequest{
			Login:    validUser.Login,
			Password: "securepassword",
		})

		assert.Error(t, err)
	})
}
