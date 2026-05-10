package http

import (
	"log/slog"
	"net/http"
	"time"

	"unityaid-back/internal/config"
	"unityaid-back/internal/http/handlers"
	"unityaid-back/internal/modules/auth"
	"unityaid-back/internal/modules/events"
	"unityaid-back/internal/modules/files"
	"unityaid-back/internal/modules/news"
	"unityaid-back/internal/modules/organizations"
	"unityaid-back/internal/modules/tasks"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RouterDeps struct {
	Config config.Config
	DB     *pgxpool.Pool
	Logger *slog.Logger
}

func NewRouter(deps RouterDeps) http.Handler {
	if deps.Config.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     deps.Config.CORSAllowedOrigins,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.Static("/uploads", deps.Config.UploadsDir)

	api := router.Group("/api/v1")

	healthHandler := handlers.NewHealthHandler(deps.DB)
	api.GET("/health", healthHandler.Health)
	api.GET("/health/db", healthHandler.Database)

	authRepository := auth.NewRepository(deps.DB)
	authService := auth.NewService(authRepository, deps.Config.JWTSecret)
	authHandler := auth.NewHandler(authService)

	authGroup := api.Group("/auth")
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/logout", auth.Middleware(authService), authHandler.Logout)
	authGroup.GET("/me", auth.Middleware(authService), authHandler.Me)

	authMiddleware := auth.Middleware(authService)

	newsRepository := news.NewRepository(deps.DB)
	newsService := news.NewService(newsRepository)
	newsHandler := news.NewHandler(newsService)

	newsGroup := api.Group("/news", authMiddleware)
	newsGroup.GET("", newsHandler.List)
	newsGroup.POST("", newsHandler.Create)
	newsGroup.GET("/:id", newsHandler.Get)
	newsGroup.PUT("/:id", newsHandler.Update)
	newsGroup.DELETE("/:id", newsHandler.Delete)

	filesHandler := files.NewHandler(deps.Config.UploadsDir)
	filesGroup := api.Group("/files", authMiddleware)
	filesGroup.POST("/news-images", filesHandler.UploadNewsImage)

	organizationsRepository := organizations.NewRepository(deps.DB)
	organizationsService := organizations.NewService(organizationsRepository)
	organizationsHandler := organizations.NewHandler(organizationsService)

	organizationsGroup := api.Group("/organizations", authMiddleware)
	organizationsGroup.GET("", organizationsHandler.List)
	organizationsGroup.POST("", organizationsHandler.Create)
	organizationsGroup.GET("/:id", organizationsHandler.Get)
	organizationsGroup.PUT("/:id", organizationsHandler.Update)
	organizationsGroup.DELETE("/:id", organizationsHandler.Delete)

	eventsRepository := events.NewRepository(deps.DB)
	eventsService := events.NewService(eventsRepository)
	eventsHandler := events.NewHandler(eventsService)

	eventsGroup := api.Group("/events", authMiddleware)
	eventsGroup.GET("", eventsHandler.List)
	eventsGroup.POST("", eventsHandler.Create)
	eventsGroup.GET("/:id", eventsHandler.Get)
	eventsGroup.PUT("/:id", eventsHandler.Update)
	eventsGroup.DELETE("/:id", eventsHandler.Delete)

	tasksRepository := tasks.NewRepository(deps.DB)
	tasksService := tasks.NewService(tasksRepository)
	tasksHandler := tasks.NewHandler(tasksService)

	tasksGroup := api.Group("/tasks", authMiddleware)
	tasksGroup.GET("", tasksHandler.List)
	tasksGroup.POST("", tasksHandler.Create)
	tasksGroup.GET("/:id", tasksHandler.Get)
	tasksGroup.PUT("/:id", tasksHandler.Update)
	tasksGroup.DELETE("/:id", tasksHandler.Delete)

	return router
}
