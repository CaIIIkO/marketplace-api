package advertisement_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"marketplace-api/internal/advertisement"
	mockad "marketplace-api/internal/advertisement/mock"
	"marketplace-api/internal/auth"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func setupHandlerTest(t *testing.T) (*gomock.Controller, *mockad.MockServiceInterface, *advertisement.Handler) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockService := mockad.NewMockServiceInterface(ctrl)
	handler := advertisement.NewAdHandler(mockService)
	return ctrl, mockService, handler
}

func withUserContext(r *http.Request, userID uuid.UUID) *http.Request {
	ctx := auth.WithUserID(r.Context(), userID)
	return r.WithContext(ctx)
}

func TestHandler_CreateAd(t *testing.T) {
	validInput := &advertisement.CreateAdvertisementInput{
		Title:        "Valid Title",
		Description:  "Valid description",
		ImageURL:     "http://example.com/image.jpg",
		PriceKopecks: 1000,
	}
	validUserID := uuid.New()
	expectedAd := &advertisement.Advertisement{
		Title:        validInput.Title,
		Description:  validInput.Description,
		ImageURL:     validInput.ImageURL,
		PriceKopecks: validInput.PriceKopecks,
		AuthorID:     validUserID,
	}

	t.Run("успешное создание", func(t *testing.T) {
		ctrl, mockService, handler := setupHandlerTest(t)
		defer ctrl.Finish()

		body, _ := json.Marshal(validInput)
		req := httptest.NewRequest(http.MethodPost, "/advertisement", bytes.NewReader(body))
		req = withUserContext(req, validUserID)
		w := httptest.NewRecorder()

		mockService.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(expectedAd, nil)

		handler.CreateAd(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result advertisement.Advertisement
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, validInput.Title, result.Title)
	})

	t.Run("ошибка: не POST", func(t *testing.T) {
		_, _, handler := setupHandlerTest(t)

		req := httptest.NewRequest(http.MethodGet, "/advertisement", nil)
		w := httptest.NewRecorder()

		handler.CreateAd(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Result().StatusCode)
	})

	t.Run("ошибка: неавторизован", func(t *testing.T) {
		_, _, handler := setupHandlerTest(t)

		body, _ := json.Marshal(validInput)
		req := httptest.NewRequest(http.MethodPost, "/advertisement", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.CreateAd(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
	})

	t.Run("ошибка: невалидный JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/advertisement", bytes.NewReader([]byte("{bad-json")))
		req = withUserContext(req, validUserID)
		w := httptest.NewRecorder()

		_, _, handler := setupHandlerTest(t)
		handler.CreateAd(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	})

	t.Run("ошибка: пустые обязательные поля", func(t *testing.T) {
		invalid := &advertisement.CreateAdvertisementInput{
			Title: "", Description: "", ImageURL: "", PriceKopecks: 0,
		}
		body, _ := json.Marshal(invalid)
		req := httptest.NewRequest(http.MethodPost, "/advertisement", bytes.NewReader(body))
		req = withUserContext(req, validUserID)
		w := httptest.NewRecorder()

		_, _, handler := setupHandlerTest(t)
		handler.CreateAd(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	})

	t.Run("ошибка: ошибка сервиса", func(t *testing.T) {
		ctrl, mockService, handler := setupHandlerTest(t)
		defer ctrl.Finish()

		body, _ := json.Marshal(validInput)
		req := httptest.NewRequest(http.MethodPost, "/advertisement", bytes.NewReader(body))
		req = withUserContext(req, validUserID)
		w := httptest.NewRecorder()

		mockService.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("something went wrong"))

		handler.CreateAd(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	})
}

func TestHandler_ListAd(t *testing.T) {
	validUserID := uuid.New()
	expectedResult := &[]advertisement.AdvertisementList{
		{
			Title:        "Test Ad",
			Description:  "A description",
			ImageURL:     "http://image.com",
			PriceKopecks: 12345,
			AuthorLogin:  "San",
		},
	}

	t.Run("успешный запрос", func(t *testing.T) {
		ctrl, mockService, handler := setupHandlerTest(t)
		defer ctrl.Finish()

		req := httptest.NewRequest(http.MethodGet, "/advertisement/?page=2&limit=5&min_price_kopecks=100&max_price_kopecks=1000&sort_by=price&sort_direction=asc", nil)
		req = withUserContext(req, validUserID)
		w := httptest.NewRecorder()

		expectedParams := &advertisement.AdvertisementListParams{
			Page:            2,
			Limit:           5,
			MinPriceKopecks: 100,
			MaxPriceKopecks: 1000,
			SortBy:          "price",
			SortDirection:   "asc",
			UserID:          &validUserID,
		}

		mockService.EXPECT().ListAd(gomock.Any(), expectedParams).Return(expectedResult, nil)

		handler.ListAd(w, req)
		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result []advertisement.Advertisement
		err := json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)
	})

	t.Run("ошибка метода", func(t *testing.T) {
		_, _, handler := setupHandlerTest(t)

		req := httptest.NewRequest(http.MethodPost, "/advertisements/", nil)
		w := httptest.NewRecorder()

		handler.ListAd(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		assert.Contains(t, w.Body.String(), "method not allowed")
	})

	t.Run("ошибка: невалидный page", func(t *testing.T) {
		_, _, handler := setupHandlerTest(t)
		req := httptest.NewRequest(http.MethodGet, "/advertisement/?page=abc", nil)
		w := httptest.NewRecorder()

		handler.ListAd(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "page must be an integer")
	})

	t.Run("ошибка: невалидный limit", func(t *testing.T) {
		_, _, handler := setupHandlerTest(t)
		req := httptest.NewRequest(http.MethodGet, "/advertisement/?limit=bad", nil)
		w := httptest.NewRecorder()

		handler.ListAd(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "limit must be an integer")
	})

	t.Run("ошибка: невалидный min_price", func(t *testing.T) {
		_, _, handler := setupHandlerTest(t)
		req := httptest.NewRequest(http.MethodGet, "/advertisement/?min_price_kopecks=err", nil)
		w := httptest.NewRecorder()

		handler.ListAd(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "min_price_kopecks must be an integer")
	})

	t.Run("ошибка: невалидный max_price", func(t *testing.T) {
		_, _, handler := setupHandlerTest(t)
		req := httptest.NewRequest(http.MethodGet, "/advertisement/?max_price_kopecks=wrong", nil)
		w := httptest.NewRecorder()

		handler.ListAd(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "max_price_kopecks must be an integer")
	})

	t.Run("ошибка: сервис вернул ошибку", func(t *testing.T) {
		ctrl, mockService, handler := setupHandlerTest(t)
		defer ctrl.Finish()

		req := httptest.NewRequest(http.MethodGet, "/advertisement/?page=1&limit=10", nil)
		req = withUserContext(req, validUserID)
		w := httptest.NewRecorder()

		expectedParams := &advertisement.AdvertisementListParams{
			Page:            1,
			Limit:           10,
			MinPriceKopecks: 0,
			MaxPriceKopecks: 0,
			UserID:          &validUserID,
		}

		mockService.EXPECT().ListAd(gomock.Any(), expectedParams).Return(nil, errors.New("service error"))

		handler.ListAd(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "service error")
	})
}
