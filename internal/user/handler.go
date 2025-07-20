package user

import (
	"context"
	"encoding/json"
	"net/http"
)

type ServiceInterface interface {
	Register(ctx context.Context, input *RegisterRequest) (*User, error)
	Authenticate(ctx context.Context, input *LoginRequest) (string, error)
}

type Handler struct {
	service ServiceInterface
}

func NewUserHandler(service ServiceInterface) *Handler {
	return &Handler{service: service}
}

// Login godoc
// @Summary Аунтификация пользователя
// @Description Принимает email и пароль, возвращает JWT-токен
// @Tags auth
// @Accept json
// @Produce json
// @Param input body LoginRequest true "Данные для аунтификации"
// @Success 200 {object} LoginResponse
// @Failure 400 {string} string "Неверный ввод"
// @Failure 401 {string} string "Неавторизован"
// @Failure 405 {string} string "Метод не разрешён"
// @Router /login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	// Проверка обязательных полей
	if input.Login == "" || input.Password == "" {
		http.Error(w, "all fields are required", http.StatusBadRequest)
		return
	}

	//Вызов сервиса
	token, err := h.service.Authenticate(r.Context(), &input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{
		Token: token,
	})
}

// Register godoc
// @Summary Регистрация нового пользователя
// @Description Принимает данные пользователя и создаёт новую учётную запись
// @Tags auth
// @Accept json
// @Produce json
// @Param input body RegisterRequest true "Данные для регистрации"
// @Success 201 {object} RegisterResponse
// @Failure 400 {string} string "Неверный ввод"
// @Failure 405 {string} string "Метод не разрешён"
// @Router /register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	// Проверка обязательных полей
	if input.Login == "" || input.Password == "" {
		http.Error(w, "all fields are required", http.StatusBadRequest)
		return
	}

	//Вызов сервиса
	user, err := h.service.Register(r.Context(), &input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(RegisterResponse{
		ID:    user.ID,
		Login: user.Login,
	})
}
