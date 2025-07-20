package advertisement_test

import (
	"context"
	"errors"
	"marketplace-api/internal/advertisement"
	mockad "marketplace-api/internal/advertisement/mock"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func setupTest(t *testing.T) (*gomock.Controller, *mockad.MockRepositoryInterface, *advertisement.Service) {
	t.Helper()

	ctrl := gomock.NewController(t)
	mockRepo := mockad.NewMockRepositoryInterface(ctrl)
	service := advertisement.NewAdService(mockRepo)

	return ctrl, mockRepo, service
}

func TestService_Create(t *testing.T) {
	validInput := &advertisement.CreateAdvertisementInput{
		Title:        "Valid title 123",
		Description:  "Description of ad",
		ImageURL:     "http://example.com/image.jpg",
		PriceKopecks: 1000,
		AuthorID:     uuid.New(),
	}
	t.Run("успешное создание", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mockad.NewMockRepositoryInterface(ctrl)
		service := advertisement.NewAdService(mockRepo)

		mockRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(&advertisement.Advertisement{
				Title:        validInput.Title,
				Description:  validInput.Description,
				ImageURL:     validInput.ImageURL,
				PriceKopecks: validInput.PriceKopecks,
				AuthorID:     validInput.AuthorID,
			}, nil)

		ad, err := service.Create(context.Background(), validInput)
		assert.NoError(t, err)
		assert.Equal(t, validInput.Title, ad.Title)
	})

	t.Run("валидация: короткий title", func(t *testing.T) {
		ctrl, _, service := setupTest(t)
		defer ctrl.Finish()

		badInput := *validInput
		badInput.Title = "a"
		_, err := service.Create(context.Background(), &badInput)
		assert.Error(t, err)
		assert.Regexp(t, regexp.MustCompile("title must be"), err.Error())
	})

	t.Run("валидация: недопустимый title", func(t *testing.T) {
		ctrl, _, service := setupTest(t)
		defer ctrl.Finish()

		badInput := *validInput
		badInput.Title = "Заголовок ^_^"
		_, err := service.Create(context.Background(), &badInput)
		assert.Error(t, err)
		assert.Regexp(t, regexp.MustCompile("title must contain"), err.Error())
	})

	t.Run("валидация: не верный URL", func(t *testing.T) {
		ctrl, _, service := setupTest(t)
		defer ctrl.Finish()

		badInput := *validInput
		badInput.ImageURL = "ftp://example.com/image.jpg"
		_, err := service.Create(context.Background(), &badInput)
		assert.Error(t, err)
		assert.Regexp(t, regexp.MustCompile("invalid image URL"), err.Error())
	})

	t.Run("валидация: не верный URL", func(t *testing.T) {
		ctrl, _, service := setupTest(t)
		defer ctrl.Finish()

		badInput := *validInput
		badInput.ImageURL = "fsdfs"
		_, err := service.Create(context.Background(), &badInput)
		assert.Error(t, err)
		assert.Regexp(t, regexp.MustCompile("invalid image URL"), err.Error())
	})

	t.Run("валидация: нет описания", func(t *testing.T) {
		ctrl, _, service := setupTest(t)
		defer ctrl.Finish()

		badInput := *validInput
		badInput.Description = ""
		_, err := service.Create(context.Background(), &badInput)
		assert.Error(t, err)
		assert.Regexp(t, regexp.MustCompile("description must contain"), err.Error())
	})

	t.Run("валидация: отрицательная цена", func(t *testing.T) {
		ctrl, _, service := setupTest(t)
		defer ctrl.Finish()

		badInput := *validInput
		badInput.PriceKopecks = -100
		_, err := service.Create(context.Background(), &badInput)
		assert.Error(t, err)
		assert.Regexp(t, regexp.MustCompile("invalid price"), err.Error())
	})

	//
	t.Run("тест ошибки из репозитория", func(t *testing.T) {
		ctrl, mockRepo, service := setupTest(t)
		defer ctrl.Finish()

		mockRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("db error"))

		_, err := service.Create(context.Background(), validInput)
		assert.Error(t, err)
		assert.Equal(t, "db error", err.Error())
	})

}

func TestService_ListAd(t *testing.T) {

	params := &advertisement.AdvertisementListParams{
		Page:            1,
		Limit:           10,
		SortBy:          "created_at",
		SortDirection:   "desc",
		MinPriceKopecks: 0,
		MaxPriceKopecks: 0,
	}

	expectedList := []advertisement.AdvertisementList{
		{
			Title:        "test",
			Description:  "test",
			ImageURL:     "https://i.pinimg.com/originals/c0/2d/11/c02d11b807f28927def41b6346cb6da0.jpg",
			PriceKopecks: 110,
			AuthorLogin:  "san"},
	}

	t.Run("успешный поиск", func(t *testing.T) {
		ctrl, mockRepo, service := setupTest(t)
		defer ctrl.Finish()

		mockRepo.EXPECT().
			GetAdvertisementsList(gomock.Any(), gomock.Eq(params)).
			Return(expectedList, nil)

		list, err := service.ListAd(context.Background(), params)
		assert.NoError(t, err)
		assert.NotNil(t, list)
		assert.Equal(t, len(expectedList), len(*list))
	})

	t.Run("валидация: minPrice > maxPrice", func(t *testing.T) {
		ctrl, _, service := setupTest(t)
		defer ctrl.Finish()

		badParams := *params
		badParams.MinPriceKopecks = 100
		badParams.MaxPriceKopecks = 10

		_, err := service.ListAd(context.Background(), &badParams)
		assert.Error(t, err)
		assert.Regexp(t, regexp.MustCompile("minimum price cannot be higher"), err.Error())
	})

	t.Run("валидация: maxPrice < 0", func(t *testing.T) {
		ctrl, mockRepo, service := setupTest(t)
		defer ctrl.Finish()

		badParams := *params
		badParams.MaxPriceKopecks = -100

		mockRepo.EXPECT().
			GetAdvertisementsList(gomock.Any(), gomock.Eq(params)).
			Return(expectedList, nil)

		list, err := service.ListAd(context.Background(), &badParams)
		assert.NoError(t, err)
		assert.NotNil(t, list)
	})

	t.Run("валидация: minPrice < 0", func(t *testing.T) {
		ctrl, mockRepo, service := setupTest(t)
		defer ctrl.Finish()

		badParams := *params
		badParams.MinPriceKopecks = -100

		mockRepo.EXPECT().
			GetAdvertisementsList(gomock.Any(), gomock.Eq(params)).
			Return(expectedList, nil)

		list, err := service.ListAd(context.Background(), &badParams)
		assert.NoError(t, err)
		assert.NotNil(t, list)
	})

	t.Run("валидация: page < 1", func(t *testing.T) {
		ctrl, mockRepo, service := setupTest(t)
		defer ctrl.Finish()

		badParams := *params
		badParams.Page = -100

		mockRepo.EXPECT().
			GetAdvertisementsList(gomock.Any(), gomock.Eq(params)).
			Return(expectedList, nil)

		list, err := service.ListAd(context.Background(), &badParams)
		assert.NoError(t, err)
		assert.NotNil(t, list)
	})

	t.Run("валидация: limit < 1", func(t *testing.T) {
		ctrl, mockRepo, service := setupTest(t)
		defer ctrl.Finish()

		badParams := *params
		badParams.Limit = -100

		mockRepo.EXPECT().
			GetAdvertisementsList(gomock.Any(), gomock.Eq(params)).
			Return(expectedList, nil)

		list, err := service.ListAd(context.Background(), &badParams)
		assert.NoError(t, err)
		assert.NotNil(t, list)
	})

	// Тест ошибки из репозитория
	t.Run("тест ошибки из репозитория", func(t *testing.T) {
		ctrl, mockRepo, service := setupTest(t)
		defer ctrl.Finish()

		mockRepo.EXPECT().
			GetAdvertisementsList(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("db failure"))

		validParams := &advertisement.AdvertisementListParams{
			Page:  1,
			Limit: 10,
		}
		_, err := service.ListAd(context.Background(), validParams)
		assert.Error(t, err)
		assert.Equal(t, "db failure", err.Error())
	})

}
