package advertisement

import (
	"time"

	"github.com/google/uuid"
)

type Advertisement struct {
	ID           uuid.UUID `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	ImageURL     string    `json:"image_url"`
	PriceKopecks int       `json:"price_kopecks"` //В копейках
	AuthorID     uuid.UUID `json:"author_id"`
	CreatedAt    time.Time `json:"created_at"`
}

type CreateAdvertisementInput struct {
	AuthorID     uuid.UUID `swaggerignore:"true"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	ImageURL     string    `json:"image_url"`
	PriceKopecks int       `json:"price_kopecks"`
}

type AdvertisementListParams struct {
	Page            int        `json:"page"`              // номер страницы
	Limit           int        `json:"limit"`             // количество на странице
	SortBy          string     `json:"sort_by"`           // "price" или "created_at"
	SortDirection   string     `json:"sort_direction"`    // "asc" - по возрастанию или "desc" - убыванию
	MinPriceKopecks int        `json:"min_price_kopecks"` // фильтр по цене в копейках от
	MaxPriceKopecks int        `json:"max_price_kopecks"` // фильтр по цене в копейках до
	UserID          *uuid.UUID `swaggerignore:"true"`
}

type AdvertisementList struct {
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	ImageURL     string  `json:"image_url"`
	PriceKopecks float64 `json:"price_kopecks"`
	AuthorLogin  string  `json:"author_login"`
	IsOwner      *bool   `json:"is_owner,omitempty"` // факт принадлежности объявления авторизованному пользователю
}
