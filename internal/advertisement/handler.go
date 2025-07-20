package advertisement

import (
	"context"
	"encoding/json"
	"marketplace-api/internal/auth"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

type ServiceInterface interface {
	Create(ctx context.Context, input *CreateAdvertisementInput) (*Advertisement, error)
	ListAd(ctx context.Context, params *AdvertisementListParams) (*[]AdvertisementList, error)
}

type Handler struct {
	service ServiceInterface
}

func NewAdHandler(service ServiceInterface) *Handler {
	return &Handler{service: service}
}

// CreateAd godoc
// @Summary Создать объявление
// @Description Создаёт новое объявление от авторизованного пользователя
// @Tags advertisement
// @Accept json
// @Produce json
// @Param input body CreateAdvertisementInput true "Данные объявления"
// @Success 201 {object} Advertisement
// @Failure 400 {string} string "Неверный ввод или обязательные поля пусты"
// @Failure 401 {string} string "Пользователь не авторизован"
// @Failure 405 {string} string "Метод не разрешён"
// @Security AuthToken
// @Router /advertisement [post]
func (h *Handler) CreateAd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	//Получения ID авторизованного пользователя
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var input CreateAdvertisementInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	// Проверка обязательных полей
	if input.Title == "" || input.Description == "" || input.ImageURL == "" || input.PriceKopecks == 0 {
		http.Error(w, "all fields are required", http.StatusBadRequest)
		return
	}

	//Вызов сервис
	input.AuthorID = userID
	ad, err := h.service.Create(r.Context(), &input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ad)
}

// ListAd godoc
// @Summary Получить список объявлений
// @Description Возвращает список объявлений с возможностью фильтрации и сортировки (если пользователь авторизован добавляет параметр is_owner к ответу)
// @Tags advertisement
// @Accept json
// @Produce json
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Количество элементов на странице" default(10)
// @Param sort_by query string false "Поле для сортировки (created_at, price)" default(created_at)
// @Param sort_direction query string false "Направление сортировки (asc, desc)" default(desc)
// @Param min_price_kopecks query int false "Минимальная цена в копейках" default(0)
// @Param max_price_kopecks query int false "Максимальная цена в копейках" default(0)
// @Success 200 {array} AdvertisementList
// @Failure 400 {string} string "Некорректные параметры запроса"
// @Failure 405 {string} string "Метод не разрешён"
// @Security AuthToken
// @Router /advertisement/ [get]
func (h *Handler) ListAd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем userID из контекста, если есть
	userID, ok := auth.UserIDFromContext(r.Context())
	var userIDPtr *uuid.UUID
	if ok {
		userIDPtr = &userID
	} else {
		userIDPtr = nil
	}

	// Получаем параметры запроса
	query := r.URL.Query()
	var (
		page, limit, minPrice, maxPrice int
		err                             error
	)

	//Получение page
	if pageStr := query.Get("page"); pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil {
			http.Error(w, "page must be an integer", http.StatusBadRequest)
			return
		}
	} else {
		page = 1 //Параметр по умолчанию
	}

	//Получение limit
	if limitStr := query.Get("limit"); limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			http.Error(w, "limit must be an integer", http.StatusBadRequest)
			return
		}
	} else {
		limit = 10 //Параметр по умолчанию
	}

	//Получение мминимальной цены
	if minStr := query.Get("min_price_kopecks"); minStr != "" {
		minPrice, err = strconv.Atoi(minStr)
		if err != nil {
			http.Error(w, "min_price_kopecks must be an integer", http.StatusBadRequest)
			return
		}
	} else {
		minPrice = 0 //Параметр по умолчанию
	}

	//Получение максимальной цены
	if maxStr := query.Get("max_price_kopecks"); maxStr != "" {
		maxPrice, err = strconv.Atoi(maxStr)
		if err != nil {
			http.Error(w, "max_price_kopecks must be an integer", http.StatusBadRequest)
			return
		}
	} else {
		maxPrice = 0 //Параметр по умолчанию
	}

	sortBy := query.Get("sort_by")
	sortDirection := query.Get("sort_direction")

	params := AdvertisementListParams{
		Page:            page,
		Limit:           limit,
		SortBy:          sortBy,
		SortDirection:   sortDirection,
		MinPriceKopecks: minPrice,
		MaxPriceKopecks: maxPrice,
		UserID:          userIDPtr,
	}

	listAd, err := h.service.ListAd(r.Context(), &params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(listAd)
}
