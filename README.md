# SPHINX

<img src="internal/assets/img/logo.svg" width="80" alt="logo" title="sphinx logo" />

An **Identity and Access Management (IAM)** API built with Go, designed to handle authentication, authorization, and OAuth 2.0 flows.

## 🚀 Features

### Authentication & User Management

- **Multi-entry Authentication**: Login via email, username, phone, or document
- **User Lifecycle**: User registration, email verification, password reset flows
- **Secure Session Management**: JWT-based access and refresh tokens with automatic rotation
- **Session Protection**: Automatic session termination on security violations
- **Asymmetric Keys for Signing**: JWKS endpoint for any party to validate tokens with key rotation
- **Concurrent Session Control**: Configurable limits on simultaneous sessions per user

### OAuth 2.0 Provider

- **Multi-client Support**: Both confidential clients (with secrets) and public clients (PKCE)
- **Consent Management**: User-controlled application permissions
- **Secure Grant Exchange**: Time-limited authorization codes with proper validation

### Role-Based Access Control (RBAC)

- **Client-specific Roles**: Each application can define its own role structure
- **Cross-application Authorization**: Centralized permission management

### Developer Experience

- **REST API**: Well-documented RESTful endpoints (via Swagger)
- **SDK Support**: Go client library (`sphinx` package) for easy integration
- **Middleware Ready**: Pre-built authentication and authorization middleware
- **Docker Support**: Development-ready containerization
- **Helm Support**: Production-ready helm chart

## 📋 Prerequisites

### Required Services

- **PostgreSQL** (16+): Primary database for persistent storage
- **Redis**: Cache and OAuth state management
- **[Hermes Server](https://github.com/kgjoner/hermes)**: Email delivery service for verification and notifications

### Development Tools

- **Go** 1.25 or later
- **Docker & Docker Compose**: For containerized development
- **Make**: For running build tasks

## 🚀 Quick Start

### Using Docker Compose (Recommended)

1. **Clone the repository**

   ```bash
   git clone https://github.com/kgjoner/sphinx.git
   cd sphinx
   ```

2. **Create .env file**

   ```bash
   cp .env.example .env
   ```

3. **Start the development environment**

   ```bash
   make dev
   ```

### Manual Setup

1. **Install dependencies**

   ```bash
   go mod tidy
   ```

2. **Copy and configure environment** (see Configuration section)

   ```bash
   cp .env.example .env
   ```

3. **Run database migrations**

   ```bash
   go run cmd/migrate/main.go
   ```

4. **Start the server**
   ```bash
   go run cmd/sphinx/main.go
   ```

## ⚙️ Configuration

### Core Settings

| Variable                  | Default                                                              | Description                                                                                                                                    |
| ------------------------- | -------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| `DATABASE_URL`            | `postgres://postgres:postgres@localhost:5432/sphinx?sslmode=disable` | PostgreSQL connection string                                                                                                                   |
| `REDIS_URL`               | `redis://localhost:6379/0`                                           | Redis connection string                                                                                                                        |
| `APP_VERSION`             | `v1.0.0`                                                             | Must be in semantic versioning format. Major version is used as path prefix.                                                                   |
| `SCHEME`                  | `http`                                                               | Server scheme (used by Swagger UI)                                                                                                             |
| `HOST`                    | `localhost:8080`                                                     | Server host and port (used by Swagger UI)                                                                                                      |
| `ROOT_APP_ID`             | `80cadd74-5ccd-41c4-9938-3c8961be04db`                               | Root application identifier. Matches the built-in "Sphinx app" created by migrations. Only change if you want a different application as root. |
| `EXTERNAL_AUTH_PROVIDERS` | —                                                                    | A JSON array with [providers settings](#external-providers).                                                                                   |

### Security Settings

| Variable                          | Default     | Description                                                                                                      |
| --------------------------------- | ----------- | ---------------------------------------------------------------------------------------------------------------- |
| `JWT_ACCESS_LIFETIME_IN_SEC`      | `900`       | Access token lifetime in seconds (15 min)                                                                        |
| `JWT_REFRESH_LIFETIME_IN_SEC`     | `172800`    | Refresh token lifetime in seconds (48 hours)                                                                     |
| `JWT_ALGORITHM`                   | `RS256`     | JWT signing algorithm. Supports `HS256` as legacy.                                                               |
| `JWT_SECRET`                      | `topsecret` | JWT signing secret (used for `HS256` only)                                                                       |
| `JWT_ENCRYPTION_KEY`              | `changeme`  | Key used to encrypt the private signing key                                                                      |
| `JWT_KEY_ROTATION_INTERVAL_HOURS` | `8760`      | Signing key rotation interval in hours (1 year). Set `0` to disable.                                             |
| `MAX_CONCURRENT_SESSIONS`         | `0`         | Max simultaneous sessions per user. `0` = unlimited.                                                             |
| `AUTH_GRANT_LIFETIME_IN_SEC`      | `300`       | OAuth authorization grant lifetime in seconds (5 min)                                                            |
| `SWAGGER_AUTH`                    | —           | A JSON for adding basic auth to Swagger in the format '{"user":"password"}'. Let it empty for no authentication. |

### Email and Client Integration

| Variable                   | Default                    | Description                                      |
| -------------------------- | -------------------------- | ------------------------------------------------ |
| `HERMES_BASE_URL`          | `http://localhost:8081/v1` | Hermes server URL                                |
| `HERMES_API_KEY`           | `topsecret`                | API key to authenticate with Hermes              |
| `CLIENT_BASE_URL`          | `http://localhost:3000`    | Base URL of the Sphinx client application        |
| `CLIENT_DATA_VERIFICATION` | `/verification`            | Client path for data verification (email, phone) |
| `CLIENT_PASSWORD_RESET`    | `/password/reset`          | Client path for password reset                   |

### Customization

| Variable            | Default               | Description                                                                           |
| ------------------- | --------------------- | ------------------------------------------------------------------------------------- |
| `APP_NAME`          | `Sphinx`              | Application name used in outgoing emails                                              |
| `APP_STYLE_URL`     | —                     | Custom email styling URL. Falls back to built-in style in `assets/style/`.            |
| `APP_LOGO_URL`      | —                     | Custom logo URL. Falls back to built-in SVG in `assets/img/`.                         |
| `SUPPORT_EMAIL`     | `support@example.com` | Contact address shown in outgoing emails                                              |
| `FALLBACK_LANGUAGE` | `pt-br`               | Default language for emails (`en-us` / `pt-br`)                                       |
| `EMAIL_TEMPLATES`   | —                     | A JSON for overwriting email templates in the format '{"language": {"template": {}}}' |

### Docker Compose & Build Variables

These variables are only needed when running via Docker Compose or using Makefile CI/CD targets.

| Variable            | Default    | Description                                                                                                |
| ------------------- | ---------- | ---------------------------------------------------------------------------------------------------------- |
| `DOCKER_REGISTRY`   | _(empty)_  | Docker registry for image push/pull. Leave empty for DockerHub. Also used by `make release`/`make deploy`. |
| `DB_USER`           | `postgres` | PostgreSQL username                                                                                        |
| `DB_PASSWORD`       | `postgres` | PostgreSQL password                                                                                        |
| `DB_NAME`           | `sphinx`   | PostgreSQL database name                                                                                   |
| `DB_PORT`           | `5432`     | PostgreSQL exposed port                                                                                    |
| `REDIS_PORT`        | `6379`     | Redis exposed port                                                                                         |
| `MAILHOG_SMTP_PORT` | `1025`     | Mailhog SMTP port                                                                                          |
| `MAILHOG_WEB_PORT`  | `8025`     | Mailhog web UI port                                                                                        |
| `HERMES_PORT`       | `8081`     | Hermes email service port                                                                                  |

### External Providers

Use `EXTERNAL_AUTH_PROVIDERS` to configure trusted external identity providers (for example Google, Facebook, or your own endpoint). If not provided, no external auth will be possible.

Expected format:

- Value must be a JSON array.
- Each object must include: `name`, `url`, `subjectIDPath`.
- `method` is optional and defaults to `GET`.
- Unknown fields are rejected during startup.

Example:

```json
[
  {
    "name": "google",
    "url": "https://openidconnect.googleapis.com/v1/userinfo",
    "method": "GET",
    "defaultHeaders": {
      "Accept": "application/json"
    },
    "defaultParams": {},
    "defaultBody": {},
    "subjectIDPath": "sub",
    "emailPath": "email",
    "aliasPath": "name"
  }
]
```

How it works:

- Sphinx sends a request to the provider URL using the configured method.
- It forwards the incoming `Authorization` header.
- It merges default headers/params/body with runtime values.
- It extracts user data from response JSON using dot-paths (for example `profile.id`).

Operational notes:

- Keep provider names unique to avoid ambiguous behavior.
- Only use trusted providers, because external auth directly affects account identity.
- If startup fails with provider parsing/validation errors, verify JSON syntax and required fields.
- If authentication fails at runtime, verify `subjectIDPath` and optional paths (`emailPath`, `aliasPath`) match the provider response payload.

## 🔗 API Integration

### Using the Go SDK

```go
import "github.com/kgjoner/sphinx/pkg/sphinx"

// Initialize the service
svc := sphinx.New("http://localhost:8080", "app-id", "app-secret")

// Use authentication middleware
router.Use(svc.Middlewares().Authenticate)
router.Use(svc.Middlewares().Guard("ADMIN"))
```

### REST API Examples

**User Registration**

```bash
POST /v1/api/user
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword",
  "username": "johndoe"
}
```

**Authentication**

```bash
POST /v1/api/auth/login
Content-Type: application/json

{
  "entry": "user@example.com",
  "password": "securepassword"
}
```

**OAuth Authorization**

```bash
POST /v1/api/oauth/authorize
Authorization: Bearer $MY_TOKEN
Content-Type: application/json

{
  "client_id": "your-app-id",
  "redirect_uri": "https://your-app.com/callback",
  "scope": "read write",
  "state": "random-state"
}
```

## 🛠️ Development

### Available Commands

```bash
make dev                            # Start development environment
make test                           # Run all tests
make test-unit                      # Run unit tests only
make test-e2e                       # Run end-to-end tests
make doc                            # Generate API documentation
make release KIND=stable            # Build and push artifacts;
                                    # you may also define a PLATFORM variable,
                                    # which acts the same as platform flag in
                                    # docker build command
```

### Project Structure

```
├── cmd/                      # Application entrypoints
├── internal/                 # Private application code
│   |── config/               # Configuration management
│   ├── domains/{name}        # Domain logic
|   |   └── {name}case        # Application logic (use cases)
|   |   └── {name}http        # Gateway logic for http
|   |   └── {name}int         # Gateway logic used as internal client
|   |   └── {name}repo        # Adapters for SQL code
|   |   |   └── queries       # SQL raw queries
│   ├── shared/               # Shared value objects, interfaces and simple domain services
|   |   └── sharedhttp        # Middlewares shared by http gateways
│   ├── pkg/                  # Generic and project specific adapters
│   └── server/               # Composition root and HTTP server setup
├── pkg/                      # Public SDK and utilities
│   └── sphinx/               # Go client SDK
├── migrations/               # Database migrations
├── build/                    # Build related code
|   ├── helm/                 # Kubernetes deployment charts
│   └── scripts/              # Shell scripts for developing and integration
└── test/                     # Test suites
```

### Running Tests

```bash
# Run the full test suite
make test

# Run specific test categories
go test ./internal/...              # Unit tests
go test ./test/e2e/...             # End-to-end tests
go test -run TestOAuth ./test/...  # Specific test patterns
```

## 🐳 Deployment

### Docker

```bash
# Build production image
docker build -f Dockerfile -t sphinx:latest .

# Run with Docker Compose
docker compose -f docker-compose.yml up -d
```

### Kubernetes (Helm)

```bash
# Install with Helm
helm install sphinx ./build/helm \
  --set image.tag=latest \
  --set secret.databaseURL="your-db-url" \
  --set secret.redisURL="your-redis-url"
```

### Environment-specific Configurations

The `helm/` directory contains:

- `values.yaml`: Default configuration

## 📊 Monitoring

### Metrics Endpoint

Prometheus metrics are available at `/v1/metrics`:

- HTTP request metrics
- Database connection pool stats
- Authentication success/failure rates
- Session management metrics

### Health Checks

- **Health**: `GET /v1/health` - Basic health status

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Run the test suite (`make test`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Code Standards

- Follow Go conventions and `gofmt` formatting
- Write tests for new features
- Update documentation for API changes
- Use conventional commit messages

### Technology Stack

- **Language**: Go 1.25+
- **Database**: PostgreSQL with migration support
- **Cache**: Redis for OAuth state management
- **HTTP Router**: Chi router with middleware support
- **Documentation**: Swagger/OpenAPI 2.0
- **Monitoring**: Prometheus metrics integration

## 🆘 Support

- **Issues**: [GitHub Issues](https://github.com/kgjoner/sphinx/issues)
- **Discussions**: [GitHub Discussions](https://github.com/kgjoner/sphinx/discussions)
- **Email**: contato@kgjoner.com.br

## 🔗 Related Projects

- **[Hermes](https://github.com/kgjoner/hermes)**: Email delivery service
- **[Cornucopia](https://github.com/kgjoner/cornucopia)**: Go utilities and helpers

---

**Sphinx** - Secure, scalable, and developer-friendly authentication for modern applications.
