package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lucas-varjao/gohtmx/internal/auth"
	gormadapter "github.com/lucas-varjao/gohtmx/internal/auth/adapter/gorm"
	"github.com/lucas-varjao/gohtmx/internal/config"
	"github.com/lucas-varjao/gohtmx/internal/email"
	"github.com/lucas-varjao/gohtmx/internal/handlers"
	"github.com/lucas-varjao/gohtmx/internal/logger"
	"github.com/lucas-varjao/gohtmx/internal/models"
	"github.com/lucas-varjao/gohtmx/internal/service"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// gracefulShutdownTimeout limits how long we wait for in-flight requests to finish.
const gracefulShutdownTimeout = 5 * time.Second

func main() {
	cfg := loadConfigOrExit()
	initLoggerFromConfig(cfg)
	logger.Info("Iniciando servidor", "port", cfg.Server.Port)

	db := connectDatabase(cfg.Database.DSN)
	migrateDatabase(db)
	ensureAdminUser(db)

	authManager, authService := initAuthStack(db, cfg)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)

	// Build server instance
	server, err := buildServer(authHandler, authManager, db)
	if err != nil {
		logger.Error("Erro ao criar servidor", "error", err)
		os.Exit(1)
	}

	if err := runServerWithGracefulShutdown(server, cfg.Server.Port); err != nil {
		os.Exit(1)
	}
}

// loadConfigOrExit loads config and initializes a fallback logger on failure.
func loadConfigOrExit() *config.Config {
	cfg, err := config.LoadConfig()
	if err != nil {
		// Initialize logger with defaults before config is loaded
		logger.Init("info", "text")
		logger.Error("Falha ao carregar as configurações", "error", err)
		os.Exit(1)
	}
	return cfg
}

// initLoggerFromConfig normalizes log settings before initialization.
func initLoggerFromConfig(cfg *config.Config) {
	logLevel := cfg.Log.Level
	if logLevel == "" {
		logLevel = "info"
	}
	logFormat := cfg.Log.Format
	if logFormat == "" {
		logFormat = "text"
	}
	logger.Init(logLevel, logFormat)
}

// connectDatabase connects to Postgres and logs success or exits on failure.
func connectDatabase(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Error("Falha ao conectar ao banco de dados", "error", err, "dsn", dsn)
		os.Exit(1)
	}
	logger.Info("Conectado ao banco de dados", "dsn", dsn)
	return db
}

// migrateDatabase runs schema migrations needed for the app.
func migrateDatabase(db *gorm.DB) {
	if err := db.AutoMigrate(&models.User{}, &models.Session{}); err != nil {
		logger.Error("Falha ao executar migrações", "error", err)
		os.Exit(1)
	}
	logger.Info("Migrações executadas com sucesso")
}

// ensureAdminUser seeds a default admin user when missing.
func ensureAdminUser(db *gorm.DB) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Falha ao gerar hash da senha do admin", "error", err)
	}

	result := db.Where(models.User{Username: "admin"}).FirstOrCreate(&models.User{
		Username:     "admin",
		Email:        "onyx.views5004@eagereverest.com",
		DisplayName:  "Administrator",
		PasswordHash: string(passwordHash),
		Role:         "admin",
	})
	if result.Error != nil {
		logger.Error("Falha ao criar usuário admin", "error", result.Error)
	}
	logger.Info("Usuário admin verificado", "rows_affected", result.RowsAffected)
}

// initAuthStack wires adapters, auth manager, and service dependencies.
func initAuthStack(db *gorm.DB, cfg *config.Config) (*auth.AuthManager, service.AuthServiceInterface) {
	userAdapter := gormadapter.NewUserAdapter(db)
	sessionAdapter := gormadapter.NewSessionAdapter(db)
	authConfig := auth.DefaultAuthConfig()
	authManager := auth.NewAuthManager(userAdapter, sessionAdapter, authConfig)
	emailService := email.NewEmailService(cfg)
	authService := service.NewAuthService(authManager, userAdapter, emailService)
	return authManager, authService
}

// runServerWithGracefulShutdown blocks until shutdown or a server error.
func runServerWithGracefulShutdown(server *http.Server, port int) error {
	serverErr := make(chan error, 1)

	// Start server in a goroutine.
	go func() {
		logger.Info("Servidor iniciado", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Channel to receive OS signals.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for either a server error or a shutdown signal.
	select {
	case err := <-serverErr:
		logger.Error("Erro no servidor", "error", err)
		return err
	case sig := <-sigChan:
		logger.Info("Sinal de shutdown recebido", "signal", sig.String())
		logger.Info("Iniciando shutdown gracioso...")

		// Create context with timeout for graceful shutdown.
		ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
		shutdownErr := server.Shutdown(ctx)
		cancel()
		if shutdownErr != nil {
			logger.Error("Erro durante shutdown gracioso", "error", shutdownErr)
			return shutdownErr
		}

		logger.Info("Shutdown gracioso concluído")
		return nil
	}
}
