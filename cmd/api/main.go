package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	"blog-api/pkg/auth"
	"blog-api/pkg/settings"
)

// func loadConfig() *Config {
// 	return &Config{
// 		ServerHost:            getEnv[string]("HOST", "localhost"),
// 		ServerPort:            getEnv[int]("PORT", 8000),

// 		DBHost:                getEnv[string]("POSTGRES_HOST", "localhost"),
// 		DBPort:                getEnv[int]("POSTGRES_PORT", 5432),
// 		DBUser:                getEnv[string]("POSTGRES_USER", noDefault),
// 		DBPassword:            getEnv[string]("POSTGRES_PASSWORD", noDefault),
// 		DBName:                getEnv[string]("POSTGRES_DB", noDefault),
// 		DBSSLMode:             getEnv[bool]("POSTGRES_SSL", true),

// 		CacheTTLMinutes:       getEnv[int]("CACHE_TTL_MINUTES", 10),
// 	}
// }

func main() {

	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables")
	}

	for _, cfg := range []settings.EnvConfigurable{
		&auth.JWTManagerConfig{},
		&auth.PasswordManagerConfig{},
	} {
		settings.LoadConfig(cfg)
	}

	// TODO: Подключиться к базе данных
	// - Создать database.Config из параметров конфигурации
	// - Вызвать database.NewPostgresDB
	// - Обработать ошибки подключения
	// - Не забыть defer db.Close()

	// TODO: Выполнить миграции базы данных
	// - Вызвать database.Migrate(db)

	// TODO: Инициализировать JWT менеджер
	// - Создать jwtManager через auth.NewJWTManager

	// TODO: Создать слои приложения
	// 1. Репозитории (передать db)
	// 2. Сервисы (передать репозитории и jwtManager)
	// 3. Хендлеры (передать сервисы)
	// 4. Middleware (передать необходимые зависимости)

	// Настраиваем маршруты
	router := chi.NewRouter()

	// TODO: Настроить middleware
	// - Добавить глобальные middleware (logging, recovery, CORS)

	// TODO: Настроить маршруты
	// Публичные эндпоинты:
	// - POST /api/register
	// - POST /api/login
	// - GET /api/posts
	// - GET /api/posts/{id}
	// - GET /api/posts/{id}/comments
	//
	// Защищенные эндпоинты (требуют JWT):
	// - POST /api/posts
	// - PUT /api/posts/{id}
	// - DELETE /api/posts/{id}
	// - POST /api/posts/{id}/comments

	// Health check эндпоинт
	router.Get(
		"/api/health",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok","service":"blog-api"}`))
		},
	)

	// TODO: Запустить HTTP сервер
	// - Сформировать адрес из конфигурации
	// - Вывести информацию о запуске
	// - Запустить сервер и обработать ошибки
}
