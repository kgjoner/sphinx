# SPHINX

<img src="internal/assets/img/logo.svg" width="80" alt="logo" title="sphinx logo" />

An **Identity and Access Management (IAM)** API built with Go, designed to handle authentication, authorization, and OAuth 2.0 flows.

## 🚀 Features

### Authentication & User Management

- **Multi-entry Authentication**: Login via email, username, phone, or document
- **User Lifecycle**: User registration, email verification, password reset flows
- **Secure Session Management**: JWT-based access and refresh tokens with automatic rotation
- **Session Protection**: Automatic session termination on security violations
- **Concurrent Session Control**: Configurable limits on simultaneous sessions per user

### OAuth 2.0 Provider

- **Base Implementation**: Authorization code grant flow
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

2. **Start the development environment**

   ```bash
   make dev
   ```

3. **Access the API**

   - API Server: `http://localhost:8080/v1/api`
   - API Documentation: `http://localhost:8080/v1`
   - Health Check: `http://localhost:8080/v1/health`

   Auxiliary tools during development:

   - Mail Delivery Check (Mailhog): `http://localhost:8025`

### Manual Setup

1. **Install dependencies**

   ```bash
   go mod tidy
   ```

2. **Configure environment** (see Configuration section)

3. **Run database migrations**

   ```bash
   go run cmd/migrate/main.go
   ```

4. **Start the server**
   ```bash
   go run cmd/sphinx/main.go
   ```

## ⚙️ Configuration

The application uses environment variables for configuration. The default values are between brackets.

- **Production environments**: Variables whose comment below starts with an asterisk MUST HAVE a value provided. Other variables MAY HAVE custom values.

- **DEV environment using Docker Compose**: Variables whose comment below starts with an asterisk MUST NOT HAVE a value other than default. Other variables MAY HAVE custom values.

### Core Settings

```bash
DATABASE_URL=postgres://...            # PostgreSQL connection string [postgres://postgres:postgres@localhost:5432/sphynx?sslmode=disable&pool_max_conns=20]
REDIS_URL=redis://...                  # Redis connection string [redis://localhost:6379/0]
HOST=localhost:8080                    # Server host and port (for Swagger) [localhost:8080]
ROOT_APP_ID=uuid                       # Root application identifier [80cadd74-5ccd-41c4-9938-3c8961be04db]
                                         # Default value matches with the "Sphinx app" created in migrations.
                                         # Only change this if you wish an application you created as root; in
                                         # majority of cases, you would like to update "Sphinx app" values.
```

### Security Settings

```bash
JWT_SECRET=your-secret-key             # JWT signing secret [topsecret]
JWT_ACCESS_LIFETIME_IN_SEC=900         # Access token lifetime [900] (15 min)
JWT_REFRESH_LIFETIME_IN_SEC=172800     # Refresh token lifetime [172800] (48 hours)
MAX_CONCURRENT_SESSIONS=0              # Session limit [0] (0 = unlimited)
AUTH_GRANT_LIFETIME_IN_SEC=300         # OAuth grant lifetime [300] (5 min)
```

### Email Integration

```bash
HERMES_BASE_URL=https://your-hermes-server.com    # Hermes server URL.
HERMES_API_KEY=your-api-key                       # Key to connect with Hermes server.
```

### Client Integration

```bash
CLIENT_BASE_URL=https://sphinx-client.com         # Sphinx client URL.
CLIENT_DATA_VERIFICATION=/verification            # Path used in Sphinx client for verify data (email, phone) [/verification]
CLIENT_PASSWORD_RESET=/password/reset             # Path used in Sphinx client to reset password [/password/reset]
```

### Customization

```bash
APP_NAME=YourApp                       # Application name for emails [Sphinx]
APP_STYLE_URL=https://...              # Custom email styling. If none is provided, default style will be used (assets/style/style.go).
APP_LOGO_URL=https://...               # Custom logo URL. If none is provided, default logo will be used (assets/img/logo.svg)
SUPPORT_EMAIL=support@yourdomain.com   # Email address sent in emails as a contact.
FALLBACK_LANGUAGE=en-us                # Default language (en-us/pt-br) [pt-br]
```

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
DOCKER_REGISTRY=docker.io make ci   # Build and push artifacts
```

### Project Structure

```
├── cmd/                   # Application entrypoints
├── internal/              # Private application code
│   ├── domains/auth/      # Authentication domain logic
│   ├── repositories/      # Data access layer
│   ├── server/            # HTTP server setup
│   └── config/            # Configuration management
├── pkg/                   # Public SDK and utilities
│   └── sphinx/            # Go client SDK
├── build/                 # Build related code
|   ├── helm/              # Kubernetes deployment charts
│   └── scripts/           # Shell scripts for developing and integration
└── test/                  # Test suites
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
docker build -f build/Dockerfile -t sphinx:latest .

# Run with Docker Compose
docker-compose -f docker-compose.yml up -d
```

### Kubernetes (Helm)

```bash
# Install with Helm
helm install sphinx ./build/helm \
  --set image.tag=latest \
  --set database.url="your-db-url" \
  --set redis.url="your-redis-url"
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
- **Readiness**: Database and Redis connectivity checks

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

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Issues**: [GitHub Issues](https://github.com/kgjoner/sphinx/issues)
- **Discussions**: [GitHub Discussions](https://github.com/kgjoner/sphinx/discussions)
- **Email**: contato@kgjoner.com.br

## 🔗 Related Projects

- **[Hermes](https://github.com/kgjoner/hermes)**: Email delivery service
- **[Cornucopia](https://github.com/kgjoner/cornucopia)**: Go utilities and helpers

---

**Sphinx** - Secure, scalable, and developer-friendly authentication for modern applications.
