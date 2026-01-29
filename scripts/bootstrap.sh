#!/bin/bash

# Bootstrap script para renomear o template GoHTMX para um novo projeto
# Uso: ./scripts/bootstrap.sh
# 
# Este script substitui:
# - O módulo Go (github.com/lucas-varjao/gohtmx) em go.mod e imports
# - O nome do projeto (gohtmx) em arquivos de configuração

set -euo pipefail

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Valores atuais do template
OLD_MODULE="github.com/lucas-varjao/gohtmx"
OLD_PROJECT_NAME="gohtmx"

# Validar que está sendo executado na raiz do projeto
if [ ! -f "go.mod" ]; then
    echo -e "${RED}Erro: go.mod não encontrado. Execute este script na raiz do projeto.${NC}"
    exit 1
fi

echo -e "${GREEN}=== Bootstrap do Template GoHTMX ===${NC}"
echo ""
echo "Este script vai renomear o módulo Go e o nome do projeto em todos os arquivos."
echo ""

# Solicitar novo módulo Go
read -p "Digite o novo módulo Go (ex: github.com/usuario/meuprojeto): " NEW_MODULE
if [ -z "$NEW_MODULE" ]; then
    echo -e "${RED}Erro: Módulo Go não pode ser vazio.${NC}"
    exit 1
fi

# Solicitar novo nome do projeto
read -p "Digite o novo nome do projeto (ex: meuprojeto): " NEW_PROJECT_NAME
if [ -z "$NEW_PROJECT_NAME" ]; then
    echo -e "${RED}Erro: Nome do projeto não pode ser vazio.${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}Resumo das alterações:${NC}"
echo "  Módulo Go: ${OLD_MODULE} -> ${NEW_MODULE}"
echo "  Nome do projeto: ${OLD_PROJECT_NAME} -> ${NEW_PROJECT_NAME}"
echo ""
read -p "Continuar? (s/N): " CONFIRM
if [[ ! "$CONFIRM" =~ ^[sS]$ ]]; then
    echo "Cancelado."
    exit 0
fi

echo ""
echo -e "${GREEN}Iniciando substituições...${NC}"

# Função para substituir em arquivo
replace_in_file() {
    local file="$1"
    local old="$2"
    local new="$3"
    
    if [ -f "$file" ]; then
        # Verificar se o arquivo contém o texto a ser substituído
        if grep -q "$old" "$file" 2>/dev/null; then
            # Usar sed com backup (cria .bak e depois remove)
            if sed -i.bak "s|${old}|${new}|g" "$file" 2>/dev/null; then
                rm -f "${file}.bak"
                echo "  ✓ $file"
                return 0
            else
                echo -e "  ${RED}✗ $file (erro)${NC}"
                return 1
            fi
        fi
    fi
    return 0
}

# 1. Substituir módulo Go em go.mod
echo ""
echo -e "${YELLOW}Atualizando módulo Go...${NC}"
replace_in_file "go.mod" "$OLD_MODULE" "$NEW_MODULE"

# 2. Substituir imports em todos os arquivos .go e .templ
echo ""
echo -e "${YELLOW}Atualizando imports em arquivos Go e templates...${NC}"
# Usar array para evitar problemas com subshells
mapfile -t go_files < <(find . -type f \( -name "*.go" -o -name "*.templ" \) ! -path "./.git/*" ! -path "./tmp/*" ! -path "./bin/*")
for file in "${go_files[@]}"; do
    replace_in_file "$file" "$OLD_MODULE" "$NEW_MODULE"
done

# 3. Substituir nome do projeto em arquivos de configuração
echo ""
echo -e "${YELLOW}Atualizando nome do projeto em arquivos de configuração...${NC}"

# package.json
replace_in_file "package.json" "\"name\": \"${OLD_PROJECT_NAME}\"" "\"name\": \"${NEW_PROJECT_NAME}\""

# Makefile - BINARY_NAME
replace_in_file "Makefile" "BINARY_NAME=${OLD_PROJECT_NAME}" "BINARY_NAME=${NEW_PROJECT_NAME}"

# Dockerfile - comentários e binário
replace_in_file "Dockerfile" "# Dockerfile for GoHTMX" "# Dockerfile for ${NEW_PROJECT_NAME}"
replace_in_file "Dockerfile" "-o ${OLD_PROJECT_NAME}" "-o ${NEW_PROJECT_NAME}"
replace_in_file "Dockerfile" "COPY --from=builder /build/${OLD_PROJECT_NAME}" "COPY --from=builder /build/${NEW_PROJECT_NAME}"
replace_in_file "Dockerfile" "ENTRYPOINT \[\"/${OLD_PROJECT_NAME}\"\]" "ENTRYPOINT [\"/${NEW_PROJECT_NAME}\"]"

# .air.toml - bin path
replace_in_file ".air.toml" "# Config file for Air live-reloading tool (GoHTMX)" "# Config file for Air live-reloading tool (${NEW_PROJECT_NAME})"
replace_in_file ".air.toml" "cmd = \"go run github.com/a-h/templ/cmd/templ@latest generate && go build -o ./tmp/${OLD_PROJECT_NAME} .\"" "cmd = \"go run github.com/a-h/templ/cmd/templ@latest generate && go build -o ./tmp/${NEW_PROJECT_NAME} .\""
replace_in_file ".air.toml" "bin = \"tmp/${OLD_PROJECT_NAME}\"" "bin = \"tmp/${NEW_PROJECT_NAME}\""

# configs/app.yml - database DSN e email
replace_in_file "configs/app.yml" "user=${OLD_PROJECT_NAME}" "user=${NEW_PROJECT_NAME}"
replace_in_file "configs/app.yml" "password=${OLD_PROJECT_NAME}" "password=${NEW_PROJECT_NAME}"
replace_in_file "configs/app.yml" "dbname=${OLD_PROJECT_NAME}" "dbname=${NEW_PROJECT_NAME}"
replace_in_file "configs/app.yml" "from_email: 'no-reply@${OLD_PROJECT_NAME}.com'" "from_email: 'no-reply@${NEW_PROJECT_NAME}.com'"
replace_in_file "configs/app.yml" "from_name: 'GoHTMX'" "from_name: '${NEW_PROJECT_NAME}'"

# docker-compose.yml - service names, network, database credentials
replace_in_file "docker-compose.yml" "# docker-compose.yml for GoHTMX" "# docker-compose.yml for ${NEW_PROJECT_NAME}"
replace_in_file "docker-compose.yml" "POSTGRES_USER: ${OLD_PROJECT_NAME}" "POSTGRES_USER: ${NEW_PROJECT_NAME}"
replace_in_file "docker-compose.yml" "POSTGRES_PASSWORD: ${OLD_PROJECT_NAME}" "POSTGRES_PASSWORD: ${NEW_PROJECT_NAME}"
replace_in_file "docker-compose.yml" "POSTGRES_DB: ${OLD_PROJECT_NAME}" "POSTGRES_DB: ${NEW_PROJECT_NAME}"
replace_in_file "docker-compose.yml" "pg_isready -U ${OLD_PROJECT_NAME} -d ${OLD_PROJECT_NAME}" "pg_isready -U ${NEW_PROJECT_NAME} -d ${NEW_PROJECT_NAME}"
replace_in_file "docker-compose.yml" "  ${OLD_PROJECT_NAME}:" "  ${NEW_PROJECT_NAME}:"
replace_in_file "docker-compose.yml" "user=${OLD_PROJECT_NAME}" "user=${NEW_PROJECT_NAME}"
replace_in_file "docker-compose.yml" "password=${OLD_PROJECT_NAME}" "password=${NEW_PROJECT_NAME}"
replace_in_file "docker-compose.yml" "dbname=${OLD_PROJECT_NAME}" "dbname=${NEW_PROJECT_NAME}"
replace_in_file "docker-compose.yml" "${OLD_PROJECT_NAME}_network" "${NEW_PROJECT_NAME}_network"

# README.md - título e referências principais
replace_in_file "README.md" "# GoHTMX" "# ${NEW_PROJECT_NAME}"
replace_in_file "README.md" "GoHTMX é um projeto" "${NEW_PROJECT_NAME} é um projeto"

# Templates - substituir "GoHTMX" em texto visível (navbar, footer)
echo ""
echo -e "${YELLOW}Atualizando templates...${NC}"
# Converter primeira letra para maiúscula
NEW_PROJECT_NAME_CAPITALIZED=$(echo "${NEW_PROJECT_NAME:0:1}" | tr '[:lower:]' '[:upper:]')${NEW_PROJECT_NAME:1}
mapfile -t templ_files < <(find . -type f \( -name "*.templ" -o -name "*_templ.go" \) ! -path "./.git/*" ! -path "./tmp/*" ! -path "./bin/*")
for file in "${templ_files[@]}"; do
    replace_in_file "$file" "\"GoHTMX\"" "\"${NEW_PROJECT_NAME_CAPITALIZED}\""
    replace_in_file "$file" "GoHTMX" "${NEW_PROJECT_NAME_CAPITALIZED}"
done

# Atualizar go.mod após mudanças (go mod tidy)
echo ""
echo -e "${YELLOW}Executando go mod tidy...${NC}"
if go mod tidy; then
    echo -e "${GREEN}✓ go mod tidy concluído${NC}"
else
    echo -e "${YELLOW}⚠ go mod tidy retornou avisos (pode ser normal)${NC}"
fi

echo ""
echo -e "${GREEN}=== Bootstrap concluído! ===${NC}"
echo ""
echo "Próximos passos:"
echo "  1. Revise as alterações: git diff"
echo "  2. Atualize o README.md com informações do seu projeto"
echo "  3. Configure as variáveis de ambiente conforme necessário"
echo "  4. Execute: go mod download"
echo ""
