package user_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"marketplace-api/internal/user"
	mockuser "marketplace-api/internal/user/mock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func setupHandlerTest(t *testing.T) (*gomock.Controller, *mockuser.MockServiceInterface, *user.Handler) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockService := mockuser.NewMockServiceInterface(ctrl)
	handler := user.NewUserHandler(mockService)
	return ctrl, mockService, handler
}

func TestHandler_Register(t *testing.T) {
	validInput := user.RegisterRequest{
		Login:    "Valid_User_123",
		Password: "strongpassword",
	}

	t.Run("успешная регистрация", func(t *testing.T) {
		ctrl, mockService, handler := setupHandlerTest(t)
		defer ctrl.Finish()

		expectedUser := &user.User{
			ID:    uuid.New(),
			Login: strings.ToLower(validInput.Login),
		}

		mockService.EXPECT().
			Register(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, req *user.RegisterRequest) (*user.User, error) {
				assert.Equal(t, validInput.Login, req.Login)
				return expectedUser, nil
			})

		body, _ := json.Marshal(validInput)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		handler.Register(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)

		var response user.RegisterResponse
		err := json.NewDecoder(rec.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser.Login, response.Login)
	})

	t.Run("ошибка: метод не разрешён", func(t *testing.T) {
		_, _, handler := setupHandlerTest(t)
		req := httptest.NewRequest(http.MethodGet, "/register", nil)
		rec := httptest.NewRecorder()

		handler.Register(rec, req)
		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
		assert.Contains(t, rec.Body.String(), "method not allowed")
	})

	t.Run("ошибка: пустые поля", func(t *testing.T) {
		_, _, handler := setupHandlerTest(t)

		body := `{"login":"","email":"","password":""}`
		req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
		rec := httptest.NewRecorder()

		handler.Register(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "all fields are required")
	})

	t.Run("ошибка: невалидный JSON", func(t *testing.T) {
		_, _, handler := setupHandlerTest(t)
		req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader("{invalid json"))
		rec := httptest.NewRecorder()

		handler.Register(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "invalid input")
	})

	t.Run("ошибка сервиса", func(t *testing.T) {
		ctrl, mockService, handler := setupHandlerTest(t)
		defer ctrl.Finish()

		mockService.EXPECT().Register(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("some error"))

		body, _ := json.Marshal(validInput)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		handler.Register(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "some error")
	})
}

func TestHandler_Login(t *testing.T) {
	validInput := user.LoginRequest{
		Login:    "Valid_User_123",
		Password: "securepass",
	}

	t.Run("успешная аутентификация", func(t *testing.T) {
		ctrl, mockService, handler := setupHandlerTest(t)
		defer ctrl.Finish()

		mockService.EXPECT().
			Authenticate(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, req *user.LoginRequest) (string, error) {
				assert.Equal(t, validInput.Login, req.Login)
				assert.Equal(t, validInput.Password, req.Password)
				return "mocked.jwt.token", nil
			})

		body, _ := json.Marshal(validInput)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		handler.Login(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp user.LoginResponse
		err := json.NewDecoder(rec.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, "mocked.jwt.token", resp.Token)
	})

	t.Run("ошибка: метод не разрешён", func(t *testing.T) {
		_, _, handler := setupHandlerTest(t)
		req := httptest.NewRequest(http.MethodGet, "/login", nil)
		rec := httptest.NewRecorder()

		handler.Login(rec, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
		assert.Contains(t, rec.Body.String(), "method not allowed")
	})

	t.Run("ошибка: пустые поля", func(t *testing.T) {
		_, _, handler := setupHandlerTest(t)

		body := `{"email":"","password":""}`
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
		rec := httptest.NewRecorder()

		handler.Login(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "all fields are required")
	})

	t.Run("ошибка: невалидный JSON", func(t *testing.T) {
		_, _, handler := setupHandlerTest(t)
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("{invalid json"))
		rec := httptest.NewRecorder()

		handler.Login(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "invalid input")
	})

	t.Run("ошибка сервиса: неверный пароль", func(t *testing.T) {
		ctrl, mockService, handler := setupHandlerTest(t)
		defer ctrl.Finish()

		mockService.EXPECT().
			Authenticate(gomock.Any(), gomock.Any()).
			Return("", errors.New("invalid credentials"))

		body, _ := json.Marshal(validInput)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		handler.Login(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "invalid credentials")
	})
}
