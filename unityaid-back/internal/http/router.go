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
	"unityaid-back/internal/modules/forms"
	"unityaid-back/internal/modules/news"
	"unityaid-back/internal/modules/organizations"
	"unityaid-back/internal/modules/tasks"
	"unityaid-back/internal/modules/timeentries"
	"unityaid-back/internal/modules/users"

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
	authService := auth.NewService(authRepository, deps.Config.JWTSecret, deps.Config.AppEnv != "production")
	authorizer := auth.NewAuthorizer(authRepository)
	authHandler := auth.NewHandler(authService)

	authGroup := api.Group("/auth")
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/refresh", authHandler.Refresh)
	authGroup.POST("/forgot-password", authHandler.ForgotPassword)
	authGroup.POST("/reset-password", authHandler.ResetPassword)
	authGroup.POST("/verify-email", authHandler.VerifyEmail)
	authGroup.POST("/logout", auth.Middleware(authService), authHandler.Logout)
	authGroup.GET("/me", auth.Middleware(authService), authHandler.Me)
	authGroup.POST("/change-password", auth.Middleware(authService), authHandler.ChangePassword)

	authMiddleware := auth.Middleware(authService)
	canManageContent := auth.RequireRoles("super_admin", "org_admin", "coordinator")
	canManageOrganizations := auth.RequireRoles("super_admin", "org_admin")

	newsRepository := news.NewRepository(deps.DB)
	newsService := news.NewService(newsRepository)
	newsHandler := news.NewHandler(newsService, deps.Config.UploadsDir, authorizer)

	newsGroup := api.Group("/news", authMiddleware)
	newsGroup.GET("", newsHandler.List)
	newsGroup.POST("", newsHandler.Create)
	newsGroup.GET("/categories", newsHandler.ListCategories)
	newsGroup.POST("/categories", newsHandler.CreateCategory)
	newsGroup.POST("/cleanup-files", newsHandler.CleanupFiles)
	newsGroup.GET("/:id", newsHandler.Get)
	newsGroup.PUT("/:id", newsHandler.Update)
	newsGroup.DELETE("/:id", newsHandler.Delete)

	filesHandler := files.NewHandler(deps.Config.UploadsDir)
	filesGroup := api.Group("/files", authMiddleware)
	filesGroup.POST("/news-images", canManageContent, filesHandler.UploadNewsImage)
	filesGroup.POST("/organization-logos", canManageOrganizations, filesHandler.UploadOrganizationLogo)

	formsRepository := forms.NewRepository(deps.DB)
	formsService := forms.NewService(formsRepository)
	formsHandler := forms.NewHandler(formsService, authorizer)
	formsGroup := api.Group("/forms", authMiddleware)
	formsGroup.GET("/:entity/create", formsHandler.CreateForm)
	formsGroup.GET("/:entity/:id/edit", formsHandler.EditForm)

	organizationsRepository := organizations.NewRepository(deps.DB)
	organizationsService := organizations.NewService(organizationsRepository)
	organizationsHandler := organizations.NewHandler(organizationsService, authorizer)

	organizationsGroup := api.Group("/organizations", authMiddleware)
	organizationsGroup.GET("", organizationsHandler.List)
	organizationsGroup.POST("", organizationsHandler.Create)
	organizationsGroup.GET("/:id", organizationsHandler.Get)
	organizationsGroup.PUT("/:id", organizationsHandler.Update)
	organizationsGroup.DELETE("/:id", organizationsHandler.Delete)
	organizationsGroup.GET("/:id/members", organizationsHandler.ListMembers)
	organizationsGroup.POST("/:id/members", organizationsHandler.AddMember)
	organizationsGroup.PATCH("/:id/members/:memberId", organizationsHandler.UpdateMember)
	organizationsGroup.DELETE("/:id/members/:memberId", organizationsHandler.DeleteMember)

	usersRepository := users.NewRepository(deps.DB)
	usersService := users.NewService(usersRepository)
	usersHandler := users.NewHandler(usersService, authorizer)

	usersGroup := api.Group("/users", authMiddleware)
	usersGroup.GET("", canManageContent, usersHandler.ListUsers)
	usersGroup.GET("/:id", canManageContent, usersHandler.GetUser)
	usersGroup.PATCH("/:id", usersHandler.UpdateUser)
	usersGroup.GET("/:id/system-roles", usersHandler.ListUserSystemRoles)
	usersGroup.PUT("/:id/system-roles", usersHandler.UpdateUserSystemRoles)
	usersGroup.DELETE("/:id/system-roles/:roleId", usersHandler.DeleteUserSystemRole)

	systemRolesGroup := api.Group("/system-roles", authMiddleware)
	systemRolesGroup.GET("", usersHandler.ListSystemRoles)

	volunteersGroup := api.Group("/volunteers", authMiddleware)
	volunteersGroup.GET("", usersHandler.ListVolunteers)
	volunteersGroup.GET("/:id", usersHandler.GetVolunteer)
	volunteersGroup.PATCH("/:id", usersHandler.UpdateVolunteer)

	skillsGroup := api.Group("/skills", authMiddleware)
	skillsGroup.GET("", usersHandler.ListSkills)
	skillsGroup.POST("", usersHandler.CreateSkill)
	skillsGroup.DELETE("/:id", usersHandler.DeleteSkill)

	eventsRepository := events.NewRepository(deps.DB)
	eventsService := events.NewService(eventsRepository)
	eventsHandler := events.NewHandler(eventsService, authorizer)

	eventsGroup := api.Group("/events", authMiddleware)
	eventsGroup.GET("", eventsHandler.List)
	eventsGroup.POST("", eventsHandler.Create)
	eventsGroup.GET("/:id", eventsHandler.Get)
	eventsGroup.PUT("/:id", eventsHandler.Update)
	eventsGroup.DELETE("/:id", eventsHandler.Delete)
	eventsGroup.GET("/:id/applications", eventsHandler.ListApplications)
	eventsGroup.POST("/:id/applications", eventsHandler.CreateApplication)
	eventsGroup.PATCH("/:id/applications/:applicationId", eventsHandler.UpdateApplication)
	eventsGroup.DELETE("/:id/applications/:applicationId", eventsHandler.DeleteApplication)
	eventsGroup.GET("/:id/attendance", eventsHandler.ListAttendance)
	eventsGroup.POST("/:id/attendance", eventsHandler.MarkAttendance)
	eventsGroup.PATCH("/:id/attendance/:attendanceId", eventsHandler.UpdateAttendance)
	eventsGroup.GET("/:id/shifts", eventsHandler.ListShifts)
	eventsGroup.POST("/:id/shifts", eventsHandler.CreateShift)
	eventsGroup.GET("/:id/feedback", eventsHandler.ListFeedback)
	eventsGroup.POST("/:id/feedback", eventsHandler.CreateFeedback)
	eventsGroup.POST("/:id/complete", eventsHandler.Complete)

	timeEntriesRepository := timeentries.NewRepository(deps.DB)
	timeEntriesService := timeentries.NewService(timeEntriesRepository, authorizer)
	timeEntriesHandler := timeentries.NewHandler(timeEntriesService, authorizer)

	timeEntriesGroup := api.Group("/time-entries", authMiddleware)
	timeEntriesGroup.GET("", timeEntriesHandler.List)
	timeEntriesGroup.POST("", timeEntriesHandler.Create)
	timeEntriesGroup.PATCH("/:id", timeEntriesHandler.Update)
	timeEntriesGroup.POST("/:id/approve", timeEntriesHandler.Approve)
	timeEntriesGroup.POST("/:id/reject", timeEntriesHandler.Reject)

	tasksRepository := tasks.NewRepository(deps.DB)
	tasksService := tasks.NewService(tasksRepository)
	tasksHandler := tasks.NewHandler(tasksService, authorizer)

	tasksGroup := api.Group("/tasks", authMiddleware)
	tasksGroup.GET("", tasksHandler.List)
	tasksGroup.POST("", tasksHandler.Create)
	tasksGroup.GET("/:id", tasksHandler.Get)
	tasksGroup.PUT("/:id", tasksHandler.Update)
	tasksGroup.DELETE("/:id", tasksHandler.Delete)
	tasksGroup.POST("/:id/assignments", tasksHandler.AddAssignment)
	tasksGroup.DELETE("/:id/assignments/:userId", tasksHandler.RemoveAssignment)
	tasksGroup.GET("/:id/comments", tasksHandler.ListComments)
	tasksGroup.POST("/:id/comments", tasksHandler.AddComment)
	tasksGroup.GET("/:id/attachments", tasksHandler.ListAttachments)
	tasksGroup.POST("/:id/attachments", tasksHandler.AddAttachment)
	tasksGroup.GET("/:id/status-history", tasksHandler.ListStatusHistory)
	tasksGroup.GET("/:id/time-entries", tasksHandler.ListTimeEntries)
	tasksGroup.POST("/:id/time-entries", tasksHandler.AddTimeEntry)
	tasksGroup.POST("/:id/approve", tasksHandler.Approve)

	return router
}
