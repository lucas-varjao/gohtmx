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

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		// Initialize logger with defaults before config is loaded
		logger.Init("info", "text")
		logger.Error("Falha ao carregar as configurações", "error", err)
		os.Exit(1)
	}

	// Initialize logger with config
	logLevel := cfg.Log.Level
	if logLevel == "" {
		logLevel = "info"
	}
	logFormat := cfg.Log.Format
	if logFormat == "" {
		logFormat = "text"
	}
	logger.Init(logLevel, logFormat)

	logger.Info("Iniciando servidor", "port", cfg.Server.Port)

	dbDSN := cfg.Database.DSN

	// Connect to PostgreSQL
	db, err := gorm.Open(postgres.Open(dbDSN), &gorm.Config{})
	if err != nil {
		logger.Error("Falha ao conectar ao banco de dados", "error", err, "dsn", dbDSN)
		os.Exit(1)
	}
	logger.Info("Conectado ao banco de dados", "dsn", dbDSN)

	// Migrate tables (including new Session table)
	if err := db.AutoMigrate(&models.User{}, &models.Session{}); err != nil {
		logger.Error("Falha ao executar migrações", "error", err)
		os.Exit(1)
	}
	logger.Info("Migrações executadas com sucesso")

	// Create admin user if not exists
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

	// Initialize adapters
	userAdapter := gormadapter.NewUserAdapter(db)
	sessionAdapter := gormadapter.NewSessionAdapter(db)

	// Initialize auth manager with default config
	authConfig := auth.DefaultAuthConfig()
	authManager := auth.NewAuthManager(userAdapter, sessionAdapter, authConfig)

	// Initialize services
	emailService := email.NewEmailService(cfg)
	authService := service.NewAuthService(authManager, userAdapter, emailService)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)

	// Build server instance
	server, err := buildServer(authHandler, authManager, db)
	if err != nil {
		logger.Error("Erro ao criar servidor", "error", err)
		os.Exit(1)
	}

	// Channel to receive server errors
	serverErr := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		logger.Info("Servidor iniciado", "port", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Channel to receive OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for either a server error or a shutdown signal
	select {
	case err := <-serverErr:
		logger.Error("Erro no servidor", "error", err)
		os.Exit(1)
	case sig := <-sigChan:
		logger.Info("Sinal de shutdown recebido", "signal", sig.String())
		logger.Info("Iniciando shutdown gracioso...")

		// Create context with timeout for graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Attempt graceful shutdown
		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Erro durante shutdown gracioso", "error", err)
			os.Exit(1)
		}

		logger.Info("Shutdown gracioso concluído")
	}
}
