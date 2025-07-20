
# 🛒 REST-API Marketplace

**Тестовое задание для VK — Стажировка Backend-разработчик**
---
# 👀 Обзор проекта
**Реализованы функции**:
- Авторизация и регистрация пользователей
- Создание объявлений
- Получение ленты объявлений с пагинацией, сортировкой, фильтрацией

# 🏗️ Используемые технологии
- `Go + net/http`
- База данных: `PostgreSQL`
- Работа с базой данных: `pgx`
- Миграции: `Goose`
- Аутентификация и авторизация: `JWT`
- Документация: `Swagger`
- `Docker + Docker Compose`

# 🗄️ Структура проекта

```bash
. marketplace-api
├── .env                     # Системные переменные окружения
├── .gitignore               # Файл игнорируемых Git-объектов
├── docker-compose.yaml      # Конфигурация для docker-compose
├── Dockerfile               # Инструкция сборки Docker-образа
├── go.mod                   
├── go.sum                   
├── Makefile                 # Набор команд для сборки, тестирования и запуска
├── Readme.md                # Основная документация проекта

├── cmd/                     # Точка входа в приложение
│   └── main.go              # Главный файл запуска HTTP-сервера

├── docs/                    # Документация Swagger
│   ├── docs.go              # Автогенерируемая документация Swagger
│   ├── swagger.json         # Swagger-документация в формате JSON
│   └── swagger.yaml         # Swagger-документация в формате YAML

├── internal/                # Внутренние пакеты
│   ├── advertisement/       # Логика работы с объявлениями
│   │   ├── handler.go              # HTTP-хендлеры
│   │   ├── handler_test.go         # Тесты для хендлеров
│   │   ├── model.go                # Модели данных объявлений
│   │   ├── repository.go           # Работа с базой данных
│   │   ├── service.go              # Бизнес-логика
│   │   ├── service_test.go         # Тесты бизнес-логики
│   │   └── mock/                   # Моки для юнит-тестов
│   │       ├── mock_repository_interface.go
│   │       └── mock_service_interface.go
│
│   ├── auth/               # Авторизация и аутентификация
│   │   ├── jwtManager.go           # Работа с JWT-токенами
│   │   └── middleware.go           # Middleware для проверки авторизации
│
│   ├── db/                 # Работа с базой данных
│   │   └── pgx.go                  # Инициализация соединения с PostgreSQL
│
│   └── user/               # Логика работы с пользователями
│       ├── handler.go             # HTTP-хендлеры
│       ├── handler_test.go        # Тесты хендлеров
│       ├── model.go               # Модели пользователей
│       ├── repository.go          # Работа с базой данных
│       ├── service.go             # Бизнес-логика
│       ├── service_test.go        # Тесты бизнес-логики
│       └── mock/                  # Моки для тестов
│           ├── mock_repository_interface.go
│           └── mock_service_interface.go

└── migrations/             # SQL-миграции для базы данных
    ├── 20250716101404_create_user.sql             
    └── 20250716120309_create_advertisement.sql     
```


## 🚀 Как запустить
1. Прогон unit-тестов (опционально)
```bash
make test
```
2. Сборка и запуск docker-контейнеров
```bash
make up
```
3. Установка утилиты для работы с миграциями
```bash
make goose-install
```
4. Накатывание миграций на базу данных
```bash
make goose-up
```

## ✳️ Swagger
Документация Swagger доступна по адресу: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)



## ⚙️ API Эндпоинты
```
http://localhost:8080
```
## 1. Регистрация
URL: `/register`

Метод: `POST`

Content-Type: `application/json`

Тело запроса:
```bash
  {
    "login": "Sanches",
    "password": "Syperpa?ssword1"
  }
```

Пример запроса:
```bash
curl -X 'POST'
  'http://localhost:8080/register'
  -H 'accept: application/json'
  -H 'Content-Type: application/json'
  -d '{
    "login": "Sanches",
    "password": "Syperpa?ssword1"
  }'
```
### Валидация логина (`login`)

- ✅ От 3 до 30 символов  
- ✅ Допустимые символы: `a-z`, `A-Z`, `0-9`, `_`, `.`  
- ❌ **Не может начинаться** с цифры, `_` или `.`  
- ❌ **Не может заканчиваться** на `_` или `.`  
- ❌ **Не может содержать** `__`, `..`, `_.`, `._`
- ℹ️ Логины **нечувствительны к регистру** — например, `VALID_LOGIN` считается эквивалентом `valid_login`


### Валидация пароля (`password`)

- ✅ Длина от 6 до 30 символов  
- ✅ Допустимые символы `a-z`, `A-Z`, `0-9`, `!@#$%^&*()_+\-=\[\]{};':"\\|,.<>/?`
- ✅ Обязательно содержит:  
  - хотя бы **одну заглавную** букву  
  - хотя бы **одну строчную** букву  
  - хотя бы **одну цифру**


## 2. Авторизация
URL: `/login`

Метод: `POST`

Content-Type: `application/json`

Тело запроса:
```bash
  {
    "login": "Sanches",
    "password": "Syperpa?ssword1"
  }
```

Пример запроса:
```bash
curl -X 'POST'
  'http://localhost:8080/login' 
  -H 'accept: application/json' 
  -H 'Content-Type: application/json'
  -d '{
    "login": "Sanches",
    "password": "Syperpa?ssword1"
  }'
```

## 3. Создание объявления
URL: `/advertisement`

Метод: `POST`

Авторизация: `Authorization: Bearer <ВАШ_ТОКЕН>`

Content-Type: `application/json`



Тело запроса:
```bash
  {
    "title": "Отдам котёнка в добрые руки",
    "description": "Отдам милого котёнка добрым хозяинам",
    "image_url": "https://i.pinimg.com/originals/c0/2d/11/c02d11b807f28927def41b6346cb6da0.jpg",
    "price_kopecks": 100
  }
```

Пример запроса:
```bash
curl -X 'POST'
  'http://localhost:8080/advertisement'
  -H 'accept: application/json' 
  -H 'Authorization: Bearer <ВАШ_ТОКЕН>' 
  -H 'Content-Type: application/json' 
  -d '{
    "title": "Отдам котёнка в добрые руки",
    "description": "Отдам милого котёнка добрым хозяинам",
    "image_url": "https://i.pinimg.com/originals/c0/2d/11/c02d11b807f28927def41b6346cb6da0.jpg",
    "price_kopecks": 100
}'
```

### Валидация объявления (`advertisement`)

ℹ️ Все параметры в объявлении обязательны

#### `title`

- ✅ Допустимые символы: `a-z`, `A-Z`, `а-я`, `А-Я`, `0-9`, пробел  
- ✅ Длина: от 3 до 30 символов

#### `description`

- ✅ Длина: от 1 до 1000 символов  
- ❌ Не может быть пустым  
- ✅ Допускаются любые символы

### `price_kopecks`

- ✅ Значение указывается в **копейках**  
- ❌ Не может быть `<= 0`

### `image_url`

- ✅ Должен начинаться с `http` или `https`  
- ✅ Должен заканчиваться на `.jpg`, `.jpeg`, `.png`

## 4. Получение ленты объявлений
URL: `/advertisement/`

Метод: `GET`

Авторизация: `Authorization: Bearer <ВАШ_ТОКЕН>` (не обязателно, при наличии добавляет is_owner в ответ)


| Параметр          | Описание                                              | Базовое значение |
|-------------------|-------------------------------------------------------|------------------|
| `page`            | Номер страницы для пагинации                           | 1                |
| `limit`           | Количество элементов на странице                       | 10               |
| `sort_direction`  | Направление сортировки: `asc` — по возрастанию, `desc` — по убыванию | `desc`           |
| `sort_by`         | Поле сортировки: `price` — по цене, `created_at` — по дате создания | `created_at`     |
| `min_price_kopecks` | Минимальная цена в копейках для фильтрации             | 0                |
| `max_price_kopecks` | Максимальная цена в копейках для фильтрации             | 0                |



Тело запроса:
```bash
  {
    "title": "Отдам котёнка в добрые руки",
    "description": "Отдам милого котёнка добрым хозяинам",
    "image_url": "https://i.pinimg.com/originals/c0/2d/11/c02d11b807f28927def41b6346cb6da0.jpg",
    "price_kopecks": 100
}
```

Пример запроса:
```bash
curl -X 'GET'
  'http://localhost:8080/advertisement/'
  -H 'accept: application/json' 
  -H 'Authorization: Bearer <ВАШ_ТОКЕН>'
```
