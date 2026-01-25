# GoHTMX

Um template fullstack pronto para uso com **autenticação baseada em sessões**, oferecendo uma alternativa moderna aos frameworks JavaScript pesados como React. Combina Golang com TEMPL para server-side rendering, HTMX para interatividade, Tailwind CSS + DaisyUI para estilização e Alpine.js para reatividade básica.

## 📋 Visão Geral

GoHTMX é um projeto base projetado para acelerar o desenvolvimento de aplicações web fullstack sem depender de frameworks JavaScript complexos. Este template vem pré-configurado com:

- **Autenticação plugável** baseada em sessões (inspirada no Lucia Auth)
- **Server-side rendering** com TEMPL (Go 1.23+)
- **Interatividade dinâmica** com HTMX
- **UI moderna** com Tailwind CSS + DaisyUI
- **Reatividade básica** com Alpine.js
- Páginas de login e registro prontas
- Página de exemplo demonstrando toda a stack

## 🎯 Filosofia do Projeto

Este template oferece uma alternativa aos frameworks JavaScript pesados:

- ✅ **Server-side rendering** para carregamento rápido
- ✅ **HTMX** para atualizações dinâmicas sem recarregar a página
- ✅ **Alpine.js** para interatividade mínima no cliente
- ✅ **Sem build step complexo** - apenas templates Go
- ✅ **Single binary** para deploy simples
- ✅ **Menos JavaScript** = menos complexidade

## 🚀 Recursos

### Backend (Golang)

- **Template Engine**: TEMPL (server-side rendering)
- **Autenticação plugável** com adapters (estilo Lucia Auth)
- Sessões armazenadas no banco de dados
- Banco de dados SQLite com GORM
- Estrutura modular e escalável
- Middleware de autenticação
- API RESTful com Gin

### Frontend

- **Templates TEMPL** para renderização server-side
- **HTMX** para interações dinâmicas
- **Tailwind CSS + DaisyUI** para UI moderna e responsiva
- **Alpine.js** para reatividade básica
- Páginas de autenticação prontas (login, registro)
- Página de exemplo demonstrando a stack completa

## 🛠️ Pré-requisitos

- Go 1.23+ (para suporte ao TEMPL)
- Docker e Docker Compose (opcional)

## 🔧 Instalação e Uso

### Clonando o template

```bash
git clone https://github.com/lucas-varjao/gohtmx.git meu-novo-projeto
cd meu-novo-projeto
```

### Execução

```bash
cd backend
go mod download
go run cmd/server/server.go
```

O servidor estará disponível em `http://localhost:8080`

### Usando Docker Compose (opcional)

```bash
docker-compose up
```

## 📁 Estrutura do Projeto

```
gohtmx/
├── backend/
│   ├── cmd/server/           # Ponto de entrada
│   ├── configs/              # Arquivos de configuração
│   └── internal/
│       ├── auth/             # Sistema de autenticação
│       │   ├── interfaces.go # UserAdapter, SessionAdapter
│       │   ├── auth_manager.go
│       │   └── adapter/gorm/ # Implementação GORM
│       ├── config/           # Gerenciamento de configuração
│       ├── handlers/         # Handlers HTTP
│       ├── middleware/       # Middlewares (auth, CORS, rate limit)
│       ├── models/           # Modelos de dados
│       ├── repository/       # Camada de repositório
│       ├── router/           # Configuração de rotas
│       ├── service/          # Lógica de negócio
│       ├── templates/        # Templates TEMPL
│       ├── static/           # Assets estáticos (CSS, JS)
│       └── validation/       # Validação de dados
```

## 🔐 Autenticação

O sistema usa **autenticação baseada em sessões** com adapters plugáveis:

```go
// Interfaces que você pode implementar para qualquer banco
type UserAdapter interface {
    FindUserByIdentifier(identifier string) (*UserData, error)
    ValidateCredentials(identifier, password string) (*UserData, error)
    // ...
}

type SessionAdapter interface {
    CreateSession(userID string, expiresAt time.Time, metadata SessionMetadata) (*Session, error)
    GetSession(sessionID string) (*Session, error)
    // ...
}
```

### Resposta de Login

```json
{
    "session_id": "abc123...",
    "expires_at": "2024-02-11T12:00:00Z",
    "user": {
        "id": "1",
        "identifier": "admin",
        "email": "admin@example.com",
        "display_name": "Administrator",
        "role": "admin"
    }
}
```

## 🎨 Stack Frontend

### TEMPL (Templates)

Templates Go para renderização server-side:

```go
// Exemplo de template
{{ define "page" }}
<div class="container">
    <h1>{{ .Title }}</h1>
    {{ template "content" . }}
</div>
{{ end }}
```

### HTMX

Para interações dinâmicas sem JavaScript complexo:

```html
<button hx-post="/api/action" hx-target="#result">
    Clique aqui
</button>
<div id="result"></div>
```

### Alpine.js

Para reatividade básica no cliente:

```html
<div x-data="{ open: false }">
    <button @click="open = !open">Toggle</button>
    <div x-show="open">Conteúdo</div>
</div>
```

### Tailwind CSS + DaisyUI

Para estilização rápida e consistente:

```html
<button class="btn btn-primary">Botão</button>
<div class="card bg-base-100 shadow-xl">Card</div>
```

## ⚙️ Configuração

Edite o arquivo `backend/configs/app.yml` para ajustar as configurações:

```yaml
server:
    port: 8080
database:
    dsn: 'gohtmx.db'
log:
    level: 'info'
    format: 'text'
```

## 🔄 Começando um Novo Projeto

1. Clone este repositório com um novo nome
2. Personalize as configurações em `backend/configs/app.yml`
3. Modifique os modelos no backend conforme necessário
4. Adapte os templates em `backend/internal/templates/`
5. Para integrar com outro banco de usuários, implemente `UserAdapter`

## 📄 Licença

Este projeto está licenciado sob a MIT License - veja o arquivo [LICENSE](LICENSE) para detalhes.

## 🤝 Contribuição

Contribuições são bem-vindas! Por favor, sinta-se à vontade para enviar um pull request.

---

Desenvolvido com ❤️ para oferecer uma alternativa simples e eficiente aos frameworks JavaScript pesados.
