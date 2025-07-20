package advertisement

import (
	"context"
	"errors"
	"net/url"
	"regexp"
	"strings"
)

var (
	allowedImageExt        = regexp.MustCompile(`(?i)\.(jpg|jpeg|png)$`)
	allowedTitleСharacters = regexp.MustCompile(`^[a-zA-Zа-яА-Я0-9 ]+$`)
)

type RepositoryInterface interface {
	Create(ctx context.Context, ad *Advertisement) (*Advertisement, error)
	GetAdvertisementsList(ctx context.Context, params *AdvertisementListParams) ([]AdvertisementList, error)
}

type Service struct {
	repo RepositoryInterface
}

func NewAdService(repo RepositoryInterface) *Service {
	return &Service{repo: repo}
}

// Create - создание объявления
func (s *Service) Create(ctx context.Context, input *CreateAdvertisementInput) (*Advertisement, error) {
	if err := s.validateCreateInput(input); err != nil {
		return nil, err
	}

	ad := &Advertisement{
		Title:        input.Title,
		Description:  input.Description,
		ImageURL:     input.ImageURL,
		PriceKopecks: input.PriceKopecks,
		AuthorID:     input.AuthorID,
	}

	ad, err := s.repo.Create(ctx, ad)
	if err != nil {
		return nil, err
	}
	return ad, err
}

// validateCreateInput проверяет корректность входных данных при создание объявления
func (s *Service) validateCreateInput(input *CreateAdvertisementInput) error {
	title := strings.TrimSpace(input.Title)
	if len(title) < 3 || len(title) > 100 {
		return errors.New("title must be 1–100 characters")
	}
	if !allowedTitleСharacters.MatchString(title) {
		return errors.New("title must contain letters or numbers")
	}
	if len(input.Description) < 1 || len(input.Description) > 1000 {
		return errors.New("description must contain 1-1000 characters")
	}
	if input.PriceKopecks <= 0 {
		return errors.New("invalid price: must be higher than 0")
	}
	if !isValidImageURL(input.ImageURL) {
		return errors.New("invalid image URL: must start with http(s) and end with .jpg/.jpeg/.png")
	}

	return nil
}

func isValidImageURL(rawURL string) bool {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	return allowedImageExt.MatchString(u.Path)
}

// ListAd - получение списка объявлений по фильтрам
func (s *Service) ListAd(ctx context.Context, params *AdvertisementListParams) (*[]AdvertisementList, error) {
	//Валидация параметров
	params, err := s.validateListAdParams(params)
	if err != nil {
		return nil, err
	}

	adList, err := s.repo.GetAdvertisementsList(ctx, params)
	if err != nil {
		return nil, err
	}

	return &adList, nil
}

// validateListAdParams проверяет корректность параметров для получения списка объявлений
func (s *Service) validateListAdParams(params *AdvertisementListParams) (*AdvertisementListParams, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 {
		params.Limit = 10
	}
	if params.SortDirection == "" || params.SortDirection != "asc" {
		params.SortDirection = "desc"
	}
	if params.SortBy == "" || params.SortBy != "price" {
		params.SortBy = "created_at"
	}
	if params.MaxPriceKopecks < 0 {
		params.MaxPriceKopecks = 0
	}
	if params.MinPriceKopecks < 0 {
		params.MinPriceKopecks = 0
	}
	if params.MaxPriceKopecks < params.MinPriceKopecks && params.MaxPriceKopecks != 0 && params.MinPriceKopecks != 0 {
		return nil, errors.New("minimum price cannot be higher than the maximum")
	}
	return params, nil
}
