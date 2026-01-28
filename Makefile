# Makefile para GoHTMX (Go + TEMPL + HTMX + Tailwind v4 + DaisyUI)
# Projeto na raiz: main.go, server.go, go.mod
# Assets com Bun (Parcel) e templates com TEMPL

# Variáveis
BINARY_NAME=gohtmx
BUILD_DIR=bin
COVERAGE_DIR=coverage
TMP_DIR=tmp

# Cores para output
GREEN=\033[0;32m
YELLOW=\033[0;33m
RED=\033[0;31m
NC=\033[0m # No Color

.PHONY: help build run run-dev test test-short test-coverage test-integration clean \
	install install-assets mod-tidy format vet lint check check-go version \
	templ-generate assets-build assets-watch assets-dev dev

.DEFAULT_GOAL := help

help: ## Mostra esta mensagem de ajuda
	@echo -e "$(GREEN)Comandos disponíveis:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'

# ---- Go ----

templ-generate: ## Gera código Go a partir dos arquivos .templ
	@echo -e "$(GREEN)Gerando templates...$(NC)"
	@templ generate
	@echo -e "$(GREEN)Templates gerados$(NC)"

build: templ-generate ## Compila o servidor (gera templ antes)
	@echo -e "$(GREEN)Compilando servidor...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo -e "$(GREEN)Build concluído: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

run: build ## Compila e executa o servidor
	@echo -e "$(GREEN)Executando servidor...$(NC)"
	@./$(BUILD_DIR)/$(BINARY_NAME)

run-dev: ## Executa o servidor em modo desenvolvimento (go run .)
	@echo -e "$(GREEN)Executando em modo dev...$(NC)"
	@go run .

dev: ## Hot reload com air (templ + assets + go)
	@command -v air >/dev/null 2>&1 || { echo -e "$(YELLOW)air não instalado. Use: go install github.com/air-verse/air@latest$(NC)"; go run . ; exit 0; }
	@air

# ---- Frontend (bun) ----

assets-build: ## Compila CSS/JS para ./static (parcel)
	@echo -e "$(GREEN)Compilando assets...$(NC)"
	@bun run build
	@echo -e "$(GREEN)Assets em ./static$(NC)"

assets-watch: ## Watch de assets (parcel watch)
	@bun run watch

assets-dev: ## Build de assets sem otimização (dev)
	@bun run dev

# ---- Testes ----

test: ## Executa todos os testes
	@echo -e "$(GREEN)Executando testes...$(NC)"
	@go test -v ./...

test-short: ## Apenas testes rápidos (sem integração)
	@echo -e "$(GREEN)Executando testes rápidos...$(NC)"
	@go test -short -v ./...

test-coverage: ## Testes com cobertura e gera HTML
	@echo -e "$(GREEN)Executando testes com cobertura...$(NC)"
	@mkdir -p $(COVERAGE_DIR)
	@go test -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	@go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo -e "$(GREEN)Cobertura: $(COVERAGE_DIR)/coverage.html$(NC)"

test-integration: ## Apenas testes de integração
	@echo -e "$(GREEN)Executando testes de integração...$(NC)"
	@go test -v ./internal/tests/integration/...

# ---- Limpeza e deps ----

clean: ## Remove binários, cobertura e cache
	@echo -e "$(YELLOW)Limpando...$(NC)"
	@rm -rf $(BUILD_DIR) $(COVERAGE_DIR) $(TMP_DIR)
	@go clean
	@echo -e "$(GREEN)Limpeza concluída$(NC)"

install: ## Baixa dependências Go
	@echo -e "$(GREEN)Instalando dependências Go...$(NC)"
	@go mod download
	@echo -e "$(GREEN)Dependências instaladas$(NC)"

install-assets: ## Instala dependências de assets (bun)
	@echo -e "$(GREEN)Instalando dependências de assets...$(NC)"
	@bun install
	@echo -e "$(GREEN)Dependências de assets instaladas$(NC)"

mod-tidy: ## go mod tidy
	@echo -e "$(GREEN)Atualizando go.mod...$(NC)"
	@go mod tidy
	@echo -e "$(GREEN)go.mod atualizado$(NC)"

# ---- Qualidade ----

format: ## gofmt -s -w
	@echo -e "$(GREEN)Formatando código...$(NC)"
	@golangci-lint fmt
	@echo -e "$(GREEN)Código formatado$(NC)"

vet: ## go vet
	@echo -e "$(GREEN)Verificando código...$(NC)"
	@go vet ./...
	@echo -e "$(GREEN)Verificação concluída$(NC)"

lint: 
	@echo -e "$(GREEN)Verificando código...$(NC)"
	@golangci-lint run --fix
	@echo -e "$(GREEN)Verificação concluída$(NC)"

check: format lint test ## Formata, verifica e testa

# ---- Utilitários ----

check-go: ## Verifica se Go está instalado
	@which go > /dev/null || (echo -e "$(RED)Go não está instalado$(NC)" && exit 1)

version: ## Versão do Go
	@echo -e "$(GREEN)Go:$(NC)" && go version
