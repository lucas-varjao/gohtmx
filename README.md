# GoHTMX

Template fullstack com **autenticação baseada em sessões** e SSR. Combina Golang + TEMPL, HTMX, Tailwind CSS v4 + DaisyUI e Alpine.js para uma stack enxuta e rápida.

## Visão geral

GoHTMX é um projeto base para aplicações web fullstack sem frameworks JavaScript pesados. Ele já vem com:

- **Autenticação plugável** por sessão (inspirada no Lucia Auth)
- **SSR** com TEMPL (Go 1.23+)
- **Interatividade** com HTMX
- **UI** com Tailwind CSS v4 + DaisyUI
- **Reatividade leve** com Alpine.js
- Páginas de login/registro e exemplo completo

## Filosofia

- SSR para carregamento rápido
- HTMX para atualizações parciais
- Menos JavaScript, menos complexidade
- Build simples e binário único

## Recursos

### Backend (Golang)

- Gin como framework HTTP
- Autenticação baseada em sessão (DB)
- PostgreSQL com GORM
- Middlewares de autenticação, CORS e rate limit
- Área admin com gestão de usuários

### Frontend

- Templates TEMPL com layouts e componentes
- HTMX para interações
- Tailwind CSS v4 + DaisyUI
- Alpine.js para estados simples

## Pré-requisitos

- Go 1.23+
- PostgreSQL
- Bun (apenas para assets)
- Docker e Docker Compose (opcional)

## Instalação e uso

### Clone

```bash
git clone https://github.com/lucas-varjao/gohtmx.git meu-novo-projeto
cd meu-novo-projeto
```

### Rodando o servidor

```bash
go mod download
go run .
```

O servidor sobe em `http://localhost:7000` (configurável).

### Desenvolvimento com hot reload (opcional)

```bash
make dev
```

### Assets (opcional)

```bash
bun install
bun run dev
```

## Estrutura do projeto

```
gohtmx/
├── main.go                # Bootstrap do app
├── server.go              # Setup do servidor e rotas
├── configs/               # Configurações (app.yml)
├── internal/
│   ├── auth/              # Sistema de autenticação
│   ├── config/            # Carregamento de config
│   ├── handlers/          # Handlers HTTP
│   ├── middleware/        # Middlewares (auth, CORS, rate limit)
│   ├── models/            # Modelos de dados
│   ├── router/            # Setup de rotas
│   ├── service/           # Lógica de negócio
│   └── validation/        # Validação
├── templates/             # Templates TEMPL
├── assets/                # Fontes de CSS/JS
└── static/                # Assets compilados
```

## Autenticação

O sistema usa **sessões armazenadas no banco** com adapters plugáveis.

- Login retorna `session_id`
- Auth via `Authorization: Bearer {session_id}` ou cookie `session_id`

Usuário admin padrão:

- `username`: `admin`
- `password`: `admin`

## Stack Frontend

### TEMPL

```go
templ Layout(title string, body templ.Component) {
	<div class="container">
		<h1>{ title }</h1>
		@body
	</div>
}
```

### HTMX

```html
<button hx-post="/api/action" hx-target="#result">Clique aqui</button>
<div id="result"></div>
```

### Alpine.js

```html
<div x-data="{ open: false }">
  <button @click="open = !open">Toggle</button>
  <div x-show="open">Conteúdo</div>
</div>
```

### Tailwind CSS + DaisyUI

```html
<button class="btn btn-primary">Botão</button>
<div class="card bg-base-100 shadow-xl">Card</div>
```

## Configuração

Edite `configs/app.yml`:

```yaml
server:
  port: 7000
database:
  dsn: 'host=localhost user=gohtmx password=gohtmx dbname=gohtmx port=5432 sslmode=disable TimeZone=UTC'
log:
  level: 'info'
  format: 'text'
```

Em produção, use `DATABASE_DSN` para sobrescrever o DSN.

## Começando um novo projeto

1. Clone este repositório com um novo nome
2. Ajuste `configs/app.yml`
3. Modifique modelos e serviços em `internal/`
4. Adapte templates em `templates/`
5. Atualize os assets em `assets/` quando necessário

## Licença

MIT. Veja `LICENSE` para detalhes.

## Contribuição

Contribuições são bem-vindas via pull request.
