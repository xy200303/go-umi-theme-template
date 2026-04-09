package routes

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"backend/internal/api/controllers"
	"backend/internal/api/middleware"
	"backend/internal/pkg/config"
	authrepo "backend/internal/repository/auth"
	filerepo "backend/internal/repository/file"
	rolerepo "backend/internal/repository/role"
	systemrepo "backend/internal/repository/system"
	userrepo "backend/internal/repository/user"
	accesssvc "backend/internal/service/access"
	adminsvc "backend/internal/service/admin"
	authsvc "backend/internal/service/auth"
	filesvc "backend/internal/service/file"
	usersvc "backend/internal/service/user"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type AppContext struct {
	Config   *config.Config
	DB       *gorm.DB
	Redis    *redis.Client
	UserRepo *userrepo.UserRepository

	CasbinService *accesssvc.CasbinService
	AuthService   *authsvc.AuthService
	UserService   *usersvc.UserService
	AdminService  *adminsvc.AdminService
	FileService   *filesvc.FileService
	SMSService    *authsvc.SMSService

	AuthController  *controllers.AuthController
	UserController  *controllers.UserController
	AdminController *controllers.AdminController
	FileController  *controllers.FileController
}

func NewAppContext(cfg *config.Config, db *gorm.DB, redis *redis.Client) (*AppContext, error) {
	userRepo := userrepo.NewUserRepository(db)
	fileRepo := filerepo.NewFileRepository(db)
	roleRepo := rolerepo.NewRoleRepository(db)
	cfgRepo := systemrepo.NewSystemConfigRepository(db)
	auditRepo := systemrepo.NewAuditLogRepository(db)
	refreshRepo := authrepo.NewRefreshTokenRepository(redis)

	casbinService, err := accesssvc.NewCasbinService(db)
	if err != nil {
		return nil, err
	}

	smsService := authsvc.NewSMSService(cfg, redis)
	fileStorageService := filesvc.NewStorageService(cfg, cfgRepo)
	fileService := filesvc.NewFileService(cfg, fileRepo, fileStorageService, cfgRepo)
	authService := authsvc.NewAuthService(cfg, userRepo, roleRepo, refreshRepo, smsService, casbinService, fileService)
	userService := usersvc.NewUserService(cfg, userRepo, smsService, casbinService, fileService)
	adminService := adminsvc.NewAdminService(userRepo, fileRepo, roleRepo, cfgRepo, auditRepo, casbinService, redis, fileService)
	fileService.StartCleanupWorker(context.Background())

	ctx := &AppContext{
		Config:        cfg,
		DB:            db,
		Redis:         redis,
		UserRepo:      userRepo,
		CasbinService: casbinService,
		AuthService:   authService,
		UserService:   userService,
		AdminService:  adminService,
		FileService:   fileService,
		SMSService:    smsService,
	}

	ctx.AuthController = controllers.NewAuthController(authService)
	ctx.UserController = controllers.NewUserController(userService)
	ctx.AdminController = controllers.NewAdminController(adminService)
	ctx.FileController = controllers.NewFileController(fileService)

	return ctx, nil
}

func SetupRouter(app *AppContext) *gin.Engine {
	gin.SetMode(app.Config.ServerMode)
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Serve local uploaded files.
	r.Static("/static/uploads", app.Config.UploadLocalPath)

	api := r.Group("/api/v1")
	{
		api.GET("/files/:id/download", app.FileController.DownloadFile)

		auth := api.Group("/auth")
		auth.Use(middleware.AuditLogMiddleware(app.AdminService))
		{
			auth.GET("/options", app.AuthController.GetAuthOptions)
			auth.POST("/sms/send", app.AuthController.SendSMSCode)
			auth.POST("/register", app.AuthController.Register)
			auth.POST("/login/password", app.AuthController.PasswordLogin)
			auth.POST("/login/sms", app.AuthController.SMSLogin)
			auth.POST("/refresh", app.AuthController.Refresh)
			auth.POST("/logout", app.AuthController.Logout)
		}

		secured := api.Group("")
		secured.Use(
			middleware.AuthMiddleware(app.Config, app.UserRepo),
			middleware.AuditLogMiddleware(app.AdminService),
			middleware.RBACMiddleware(app.CasbinService),
		)
		{
			user := secured.Group("/user")
			{
				user.GET("/profile", app.UserController.GetProfile)
				user.PUT("/profile", app.UserController.UpdateProfile)
				user.POST("/password/reset", app.UserController.ResetPassword)
				user.POST("/phone/change", app.UserController.ChangePhone)
				user.POST("/avatar/upload", app.UserController.UploadAvatar)
				user.POST("/files/upload", app.FileController.UploadFile)
				user.POST("/files/direct/init", app.FileController.InitDirectUpload)
				user.POST("/files/direct/complete", app.FileController.CompleteDirectUpload)
			}

			admin := secured.Group("/admin")
			{
				admin.GET("/stats", app.AdminController.Stats)
				admin.GET("/files", app.AdminController.ListAdminFiles)
				admin.GET("/files/stats", app.AdminController.GetAdminFileStats)
				admin.GET("/audit-logs", app.AdminController.ListAuditLogs)
				admin.GET("/policy-templates", app.AdminController.ListPolicyTemplates)
				admin.GET("/users", app.AdminController.ListUsers)
				admin.POST("/users", app.AdminController.CreateUser)
				admin.PUT("/users/:id", app.AdminController.UpdateUser)
				admin.DELETE("/users/:id", app.AdminController.DeleteUser)
				admin.PUT("/users/:id/password", app.AdminController.ResetUserPassword)
				admin.PUT("/users/:id/roles", app.AdminController.UpdateUserRoles)
				admin.GET("/roles", app.AdminController.ListRoles)
				admin.POST("/roles", app.AdminController.CreateRole)
				admin.PUT("/roles/:id", app.AdminController.UpdateRole)
				admin.DELETE("/roles/:id", app.AdminController.DeleteRole)
				admin.GET("/roles/:id/policies", app.AdminController.GetRolePolicies)
				admin.PUT("/roles/:id/policies", app.AdminController.SetRolePolicies)
				admin.GET("/system-configs", app.AdminController.ListSystemConfigs)
				admin.PUT("/system-configs", app.AdminController.UpsertSystemConfig)
			}
		}
	}

	registerStaticWeb(r, app.Config.FrontendDistDir)
	return r
}

func registerStaticWeb(r *gin.Engine, distDir string) {
	indexPath := filepath.Join(distDir, "index.html")
	r.NoRoute(func(c *gin.Context) {
		requestPath := c.Request.URL.Path
		if requestPath != "" && requestPath != "/" {
			relativePath := strings.TrimPrefix(requestPath, "/")
			staticPath := filepath.Clean(filepath.Join(distDir, relativePath))
			distAbs, err1 := filepath.Abs(distDir)
			fileAbs, err2 := filepath.Abs(staticPath)
			if err1 == nil && err2 == nil {
				prefix := distAbs + string(os.PathSeparator)
				if fileAbs == distAbs || strings.HasPrefix(fileAbs, prefix) {
					if fi, err := os.Stat(fileAbs); err == nil && !fi.IsDir() {
						c.File(fileAbs)
						return
					}
				}
			}
		}

		if _, err := os.Stat(indexPath); err == nil {
			c.File(indexPath)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "not found"})
	})
}
