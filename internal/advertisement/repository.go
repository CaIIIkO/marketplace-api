package advertisement

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewAdRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create - создаёт объявление
func (r *Repository) Create(ctx context.Context, ad *Advertisement) (*Advertisement, error) {
	query := `
		INSERT INTO advertisements (title, description, image_url, price_kopecks, author_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	err := r.pool.QueryRow(ctx, query, ad.Title, ad.Description, ad.ImageURL, ad.PriceKopecks, ad.AuthorID).Scan(&ad.ID, &ad.CreatedAt)
	if err != nil {
		return nil, err
	}
	return ad, nil
}

// Create - получение списка объявлений по заданным параметрам
func (r *Repository) GetAdvertisementsList(ctx context.Context, params *AdvertisementListParams) ([]AdvertisementList, error) {
	offset := (params.Page - 1) * params.Limit

	//тип сортировки по полю
	var orderBy string
	switch params.SortBy {
	case "price":
		orderBy = "a.price_kopecks"
	case "created_at":
		orderBy = "a.created_at"
	default:
		orderBy = "a.created_at" // значение по умолчанию
	}

	//Направление сортировки
	sortDirection := "DESC" // значение по умолчанию
	if strings.ToUpper(params.SortDirection) == "ASC" {
		sortDirection = "ASC"
	}

	query := fmt.Sprintf(`
			SELECT 
				a.title,
				a.description,
				a.image_url,
				a.price_kopecks,
				u.login,
				CASE
					WHEN $5::uuid IS NULL THEN NULL
					WHEN a.author_id = $5 THEN true
					ELSE false
				END AS is_owner
			FROM advertisements a
			JOIN users u ON a.author_id = u.id
			WHERE ($3 = 0 OR a.price_kopecks >= $3)
			  AND ($4 = 0 OR a.price_kopecks <= $4)
			ORDER BY %s %s
			LIMIT $1 OFFSET $2`, orderBy, sortDirection)

	rows, err := r.pool.Query(ctx, query, params.Limit, offset, params.MinPriceKopecks, params.MaxPriceKopecks, params.UserID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ads []AdvertisementList

	// Если пользователь авторизован
	for rows.Next() {
		var ad AdvertisementList
		err := rows.Scan(&ad.Title, &ad.Description, &ad.ImageURL, &ad.PriceKopecks, &ad.AuthorLogin, &ad.IsOwner)
		if err != nil {
			return nil, err
		}
		ads = append(ads, ad)
	}

	return ads, nil
}
