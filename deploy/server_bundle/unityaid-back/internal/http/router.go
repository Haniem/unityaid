package http

import (
	"log/slog"
	"net/http"
	"time"

	"unityaid-back/internal/config"
	"unityaid-back/internal/http/handlers"
	"unityaid-back/internal/modules/admin"
	"unityaid-back/internal/modules/analytics"
	"unityaid-back/internal/modules/audit"
	"unityaid-back/internal/modules/auth"
	"unityaid-back/internal/modules/certificates"
	"unityaid-back/internal/modules/dataexports"
	"unityaid-back/internal/modules/events"
	"unityaid-back/internal/modules/fieldops"
	"unityaid-back/internal/modules/files"
	"unityaid-back/internal/modules/forms"
	"unityaid-back/internal/modules/gamification"
	"unityaid-back/internal/modules/invitations"
	"unityaid-back/internal/modules/knowledge"
	"unityaid-back/internal/modules/news"
	"unityaid-back/internal/modules/notifications"
	"unityaid-back/internal/modules/organizations"
	"unityaid-back/internal/modules/profilefields"
	"unityaid-back/internal/modules/shop"
	"unityaid-back/internal/modules/tasks"
	"unityaid-back/internal/modules/tenantsettings"
	"unityaid-back/internal/modules/timeentries"
	"unityaid-back/internal/modules/users"
	"unityaid-back/internal/modules/volunteerimports"

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
	authMiddleware := auth.Middleware(authService)
	auditRepository := audit.NewRepository(deps.DB)
	auditMiddleware := audit.Middleware(auditRepository)
	notificationsRepository := notifications.NewRepository(deps.DB)
	notificationsService := notifications.NewService(notificationsRepository)
	notificationsHandler := notifications.NewHandler(notificationsService)
	adminRepository := admin.NewRepository(deps.DB)
	adminService := admin.NewService(adminRepository)
	adminHandler := admin.NewHandler(adminService, authorizer)

	authGroup := api.Group("/auth")
	authLimiter := RateLimit(10, time.Minute)
	authGroup.POST("/register", authLimiter, authHandler.Register)
	authGroup.POST("/login", authLimiter, authHandler.Login)
	authGroup.POST("/refresh", authLimiter, authHandler.Refresh)
	authGroup.POST("/forgot-password", authLimiter, authHandler.ForgotPassword)
	authGroup.POST("/reset-password", authLimiter, authHandler.ResetPassword)
	authGroup.POST("/verify-email", authLimiter, authHandler.VerifyEmail)
	authGroup.POST("/logout", authMiddleware, auditMiddleware, authHandler.Logout)
	authGroup.GET("/me", authMiddleware, authHandler.Me)
	authGroup.POST("/change-password", authMiddleware, auditMiddleware, authHandler.ChangePassword)

	canManageContent := auth.RequireRoles("super_admin", "org_admin", "coordinator")
	canManageOrganizations := auth.RequireRoles("super_admin", "org_admin")

	tenantSettingsRepository := tenantsettings.NewRepository(deps.DB)
	tenantSettingsService := tenantsettings.NewService(tenantSettingsRepository)
	tenantSettingsHandler := tenantsettings.NewHandler(tenantSettingsService)
	tenantSettingsGroup := api.Group("/tenant-settings", authMiddleware, auditMiddleware)
	tenantSettingsGroup.GET("", tenantSettingsHandler.Get)
	tenantSettingsGroup.PUT("", canManageOrganizations, tenantSettingsHandler.Update)

	newsRepository := news.NewRepository(deps.DB)
	newsService := news.NewService(newsRepository)
	newsHandler := news.NewHandler(newsService, deps.Config.UploadsDir, authorizer)

	newsGroup := api.Group("/news", authMiddleware, auditMiddleware)
	newsGroup.GET("", newsHandler.List)
	newsGroup.POST("", newsHandler.Create)
	newsGroup.GET("/categories", newsHandler.ListCategories)
	newsGroup.POST("/categories", newsHandler.CreateCategory)
	newsGroup.POST("/cleanup-files", newsHandler.CleanupFiles)
	newsGroup.GET("/:id", newsHandler.Get)
	newsGroup.PUT("/:id", newsHandler.Update)
	newsGroup.DELETE("/:id", newsHandler.Delete)

	filesHandler := files.NewHandler(deps.Config.UploadsDir)
	filesGroup := api.Group("/files", authMiddleware, auditMiddleware)
	filesGroup.POST("/news-images", canManageContent, filesHandler.UploadNewsImage)
	filesGroup.POST("/organization-logos", canManageOrganizations, filesHandler.UploadOrganizationLogo)
	filesGroup.POST("/profile-avatars", filesHandler.UploadProfileAvatar)

	formsRepository := forms.NewRepository(deps.DB)
	formsService := forms.NewService(formsRepository)
	formsHandler := forms.NewHandler(formsService, authorizer)
	formsGroup := api.Group("/forms", authMiddleware, auditMiddleware)
	formsGroup.GET("/:entity/create", formsHandler.CreateForm)
	formsGroup.GET("/:entity/:id/edit", formsHandler.EditForm)

	organizationsRepository := organizations.NewRepository(deps.DB)
	organizationsService := organizations.NewService(organizationsRepository)
	organizationsHandler := organizations.NewHandler(organizationsService, authorizer)

	organizationsGroup := api.Group("/organizations", authMiddleware, auditMiddleware)
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

	usersGroup := api.Group("/users", authMiddleware, auditMiddleware)
	usersGroup.GET("", canManageContent, usersHandler.ListUsers)
	usersGroup.GET("/:id", canManageContent, usersHandler.GetUser)
	usersGroup.PATCH("/:id", usersHandler.UpdateUser)
	usersGroup.GET("/:id/system-roles", usersHandler.ListUserSystemRoles)
	usersGroup.PUT("/:id/system-roles", usersHandler.UpdateUserSystemRoles)
	usersGroup.DELETE("/:id/system-roles/:roleId", usersHandler.DeleteUserSystemRole)

	systemRolesGroup := api.Group("/system-roles", authMiddleware, auditMiddleware)
	systemRolesGroup.GET("", usersHandler.ListSystemRoles)

	volunteersGroup := api.Group("/volunteers", authMiddleware, auditMiddleware)
	volunteersGroup.GET("", usersHandler.ListVolunteers)
	volunteersGroup.GET("/:id", usersHandler.GetVolunteer)
	volunteersGroup.PATCH("/:id", usersHandler.UpdateVolunteer)

	profileFieldsRepository := profilefields.NewRepository(deps.DB)
	profileFieldsService := profilefields.NewService(profileFieldsRepository)
	profileFieldsHandler := profilefields.NewHandler(profileFieldsService, authorizer)
	profileFieldsGroup := api.Group("/profile-fields", authMiddleware, auditMiddleware)
	profileFieldsGroup.GET("/schema", profileFieldsHandler.Schema)
	profileFieldsGroup.GET("/users/:userId", profileFieldsHandler.Values)
	profileFieldsGroup.PUT("/users/:userId", profileFieldsHandler.SaveValues)
	profileFieldsGroup.POST("/groups", profileFieldsHandler.CreateGroup)
	profileFieldsGroup.PUT("/groups/:id", profileFieldsHandler.UpdateGroup)
	profileFieldsGroup.DELETE("/groups/:id", profileFieldsHandler.DeleteGroup)
	profileFieldsGroup.POST("/fields", profileFieldsHandler.CreateField)
	profileFieldsGroup.PUT("/fields/:id", profileFieldsHandler.UpdateField)
	profileFieldsGroup.DELETE("/fields/:id", profileFieldsHandler.DeleteField)

	invitationsRepository := invitations.NewRepository(deps.DB)
	invitationsService := invitations.NewService(invitationsRepository)
	invitationsHandler := invitations.NewHandler(invitationsService)
	invitationsGroup := api.Group("/invitations", authMiddleware, auditMiddleware)
	invitationsGroup.GET("", canManageOrganizations, invitationsHandler.List)
	invitationsGroup.POST("", canManageOrganizations, invitationsHandler.Create)
	invitationsGroup.POST("/:token/accept", invitationsHandler.Accept)

	volunteerImportsHandler := volunteerimports.NewHandler(deps.DB)
	volunteerImportsGroup := api.Group("/volunteer-imports", authMiddleware, auditMiddleware, canManageContent)
	volunteerImportsGroup.POST("/preview", volunteerImportsHandler.Preview)
	volunteerImportsGroup.POST("/commit", volunteerImportsHandler.Commit)

	skillsGroup := api.Group("/skills", authMiddleware, auditMiddleware)
	skillsGroup.GET("", usersHandler.ListSkills)
	skillsGroup.POST("", usersHandler.CreateSkill)
	skillsGroup.DELETE("/:id", usersHandler.DeleteSkill)

	eventsRepository := events.NewRepository(deps.DB)
	eventsService := events.NewService(eventsRepository, notificationsService)
	eventsHandler := events.NewHandler(eventsService, authorizer)

	eventsGroup := api.Group("/events", authMiddleware, auditMiddleware)
	eventsGroup.GET("", eventsHandler.List)
	eventsGroup.POST("", eventsHandler.Create)
	eventsGroup.POST("/recurring", eventsHandler.CreateRecurring)
	eventsGroup.GET("/:id", eventsHandler.Get)
	eventsGroup.PUT("/:id", eventsHandler.Update)
	eventsGroup.DELETE("/:id", eventsHandler.Delete)
	eventsGroup.GET("/:id/applications", eventsHandler.ListApplications)
	eventsGroup.POST("/:id/applications", eventsHandler.CreateApplication)
	eventsGroup.POST("/:id/applications/bulk-status", eventsHandler.BulkUpdateApplications)
	eventsGroup.PATCH("/:id/applications/:applicationId", eventsHandler.UpdateApplication)
	eventsGroup.DELETE("/:id/applications/:applicationId", eventsHandler.DeleteApplication)
	eventsGroup.GET("/:id/attendance", eventsHandler.ListAttendance)
	eventsGroup.POST("/:id/attendance", eventsHandler.MarkAttendance)
	eventsGroup.POST("/:id/attendance/bulk", eventsHandler.BulkMarkAttendance)
	eventsGroup.PATCH("/:id/attendance/:attendanceId", eventsHandler.UpdateAttendance)
	eventsGroup.GET("/:id/shifts", eventsHandler.ListShifts)
	eventsGroup.POST("/:id/shifts", eventsHandler.CreateShift)
	eventsGroup.GET("/:id/feedback", eventsHandler.ListFeedback)
	eventsGroup.POST("/:id/feedback", eventsHandler.CreateFeedback)
	eventsGroup.POST("/:id/complete", eventsHandler.Complete)

	eventTemplatesGroup := api.Group("/event-templates", authMiddleware, auditMiddleware)
	eventTemplatesGroup.GET("", eventsHandler.ListTemplates)
	eventTemplatesGroup.POST("", eventsHandler.CreateTemplate)

	fieldOpsRepository := fieldops.NewRepository(deps.DB)
	fieldOpsService := fieldops.NewService(fieldOpsRepository)
	fieldOpsHandler := fieldops.NewHandler(fieldOpsService)

	fieldOpsGroup := api.Group("/field-ops", authMiddleware, auditMiddleware)
	fieldOpsGroup.POST("/events/:eventId/qr-tokens", canManageContent, fieldOpsHandler.CreateQRToken)
	fieldOpsGroup.POST("/qr-scan", fieldOpsHandler.Scan)

	timeEntriesRepository := timeentries.NewRepository(deps.DB)
	timeEntriesService := timeentries.NewService(timeEntriesRepository, authorizer)
	timeEntriesHandler := timeentries.NewHandler(timeEntriesService, authorizer)

	timeEntriesGroup := api.Group("/time-entries", authMiddleware, auditMiddleware)
	timeEntriesGroup.GET("", timeEntriesHandler.List)
	timeEntriesGroup.POST("", timeEntriesHandler.Create)
	timeEntriesGroup.POST("/bulk-approve", timeEntriesHandler.BulkApprove)
	timeEntriesGroup.PATCH("/:id", timeEntriesHandler.Update)
	timeEntriesGroup.POST("/:id/approve", timeEntriesHandler.Approve)
	timeEntriesGroup.POST("/:id/reject", timeEntriesHandler.Reject)

	gamificationRepository := gamification.NewRepository(deps.DB)
	gamificationService := gamification.NewService(gamificationRepository, notificationsService)
	gamificationHandler := gamification.NewHandler(gamificationService, authorizer)

	gamificationGroup := api.Group("/gamification", authMiddleware, auditMiddleware)
	gamificationGroup.GET("/me", gamificationHandler.Me)
	gamificationGroup.GET("/users/:userId", gamificationHandler.UserProfile)
	gamificationGroup.GET("/leaderboard", gamificationHandler.Leaderboard)

	shopRepository := shop.NewRepository(deps.DB)
	shopService := shop.NewService(shopRepository)
	shopHandler := shop.NewHandler(shopService, authorizer)

	shopGroup := api.Group("/shop", authMiddleware, auditMiddleware)
	shopGroup.GET("/wallet", shopHandler.Wallet)
	shopGroup.POST("/wallet/transfers", shopHandler.Transfer)
	shopGroup.GET("/products", shopHandler.Products)
	shopGroup.GET("/orders", shopHandler.Orders)
	shopGroup.POST("/orders", shopHandler.CreateOrder)
	shopGroup.PATCH("/orders/:id", shopHandler.UpdateOrderStatus)

	achievementsGroup := api.Group("/achievements", authMiddleware, auditMiddleware)
	achievementsGroup.GET("", gamificationHandler.ListAchievements)
	achievementsGroup.POST("/recalculate", gamificationHandler.Recalculate)

	certificatesRepository := certificates.NewRepository(deps.DB)
	certificatesService := certificates.NewService(certificatesRepository)
	certificatesHandler := certificates.NewHandler(certificatesService, authorizer)

	certificatesGroup := api.Group("/certificates", authMiddleware, auditMiddleware)
	certificatesGroup.GET("", certificatesHandler.List)
	certificatesGroup.POST("/generate", certificatesHandler.Generate)
	certificatesGroup.GET("/download/:id", certificatesHandler.Download)
	api.GET("/certificates/verify/:code", certificatesHandler.Verify)

	analyticsRepository := analytics.NewRepository(deps.DB)
	analyticsService := analytics.NewService(analyticsRepository)
	analyticsHandler := analytics.NewHandler(analyticsService)

	analyticsGroup := api.Group("/analytics", authMiddleware, auditMiddleware)
	analyticsGroup.GET("/overview", analyticsHandler.Overview)
	analyticsGroup.GET("/volunteers", analyticsHandler.Volunteers)
	analyticsGroup.GET("/events", analyticsHandler.Events)
	analyticsGroup.GET("/tasks", analyticsHandler.Tasks)
	analyticsGroup.GET("/gamification", analyticsHandler.Gamification)
	analyticsGroup.GET("/audit", analyticsHandler.Audit)
	analyticsGroup.GET("/management/:kind", analyticsHandler.Management)
	analyticsGroup.GET("/management/:kind/export/:format", analyticsHandler.ExportManagement)

	exportsHandler := dataexports.NewHandler(deps.DB)
	exportsGroup := api.Group("/exports", authMiddleware, auditMiddleware, canManageContent)
	exportsGroup.GET("/:kind", exportsHandler.Download)

	knowledgeRepository := knowledge.NewRepository(deps.DB)
	knowledgeService := knowledge.NewService(knowledgeRepository)
	knowledgeHandler := knowledge.NewHandler(knowledgeService)

	knowledgeGroup := api.Group("/knowledge-base", authMiddleware, auditMiddleware)
	knowledgeGroup.GET("", knowledgeHandler.List)
	knowledgeGroup.POST("", canManageContent, knowledgeHandler.Create)
	knowledgeGroup.GET("/categories", knowledgeHandler.ListCategories)
	knowledgeGroup.POST("/categories", canManageContent, knowledgeHandler.CreateCategory)
	knowledgeGroup.GET("/:id", knowledgeHandler.Get)
	knowledgeGroup.PUT("/:id", canManageContent, knowledgeHandler.Update)
	knowledgeGroup.DELETE("/:id", canManageContent, knowledgeHandler.Delete)

	tasksRepository := tasks.NewRepository(deps.DB)
	tasksService := tasks.NewService(tasksRepository, notificationsService)
	tasksHandler := tasks.NewHandler(tasksService, authorizer)

	tasksGroup := api.Group("/tasks", authMiddleware, auditMiddleware)
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

	notificationsGroup := api.Group("/notifications", authMiddleware, auditMiddleware)
	notificationsGroup.GET("", notificationsHandler.List)
	notificationsGroup.POST("/:id/read", notificationsHandler.MarkRead)
	notificationsGroup.POST("/read-all", notificationsHandler.MarkAllRead)

	adminGroup := api.Group("/admin", authMiddleware, auditMiddleware, adminHandler.RequireSuperAdmin)
	adminGroup.GET("/entities", adminHandler.Entities)
	adminGroup.GET("/:entity", adminHandler.List)
	adminGroup.POST("/:entity", adminHandler.Create)
	adminGroup.PUT("/:entity/:id", adminHandler.Update)
	adminGroup.DELETE("/:entity/:id", adminHandler.Delete)

	return router
}
