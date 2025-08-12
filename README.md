# Product Hunt Clone

A modern Product Hunt clone built with Go, HTMX, and Tailwind CSS. This application demonstrates a hypermedia-driven approach to web development with server-side rendering and dynamic interactions.

## Features

- **Authentication**: OAuth integration with Yandex
- **Product Management**: Create, edit, and manage products
- **Launch System**: Submit and moderate product launches
- **User Profiles**: User management and profiles
- **Real-time Updates**: HTMX-powered dynamic interactions
- **Responsive Design**: Mobile-first design with Tailwind CSS
- **Admin Panel**: Moderation tools for launches

## Tech Stack

- **Backend**: Go 1.23+
- **Web Framework**: Chi router
- **Database**: SQLite with WAL mode
- **Frontend**: HTMX + Tailwind CSS
- **Authentication**: OAuth 2.0 (Yandex)
- **Notifications**: Telegram bot integration

## Project Structure

```
go-htmx/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── app/                    # Application services
│   ├── domain/                 # Domain models and business logic
│   ├── infra/                  # Infrastructure (database, external services)
│   ├── pkg/                    # Shared packages
│   │   ├── config/             # Configuration management
│   │   ├── errors/             # Error handling
│   │   ├── migrations/         # Database migrations
│   │   └── validation/         # Input validation
│   └── server/                 # HTTP server and handlers
│       ├── handler/            # HTTP handlers
│       └── middleware/         # HTTP middleware
├── migrations/                 # Database migration files
├── views/                      # HTML templates
└── assets/                     # Static assets (CSS, JS)
```

## Getting Started

### Prerequisites

- Go 1.23 or higher
- SQLite 3
- Yandex OAuth application (for authentication)

### Installation

1. **Clone the repository**

   ```bash
   git clone <repository-url>
   cd go-htmx
   ```

2. **Install dependencies**

   ```bash
   go mod download
   ```

3. **Set up environment variables**

   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Run database migrations**

   ```bash
   # The application will automatically run migrations on startup
   ```

5. **Start the application**
   ```bash
   go run cmd/main.go
   ```

The application will be available at `http://localhost:3333`

### Environment Variables

Create a `.env` file with the following variables:

```env
# Database
DATABASE_URL=file:./data.db

# Authentication
YANDEX_CLIENT_ID=your_yandex_client_id
YANDEX_CLIENT_SECRET=your_yandex_client_secret
SESSION_SECRET=your_session_secret_key

# Telegram (optional)
TELEGRAM_BOT_TOKEN=your_telegram_bot_token
TELEGRAM_CHAT_ID=your_telegram_chat_id

# Application
ENVIRONMENT=development
BASE_URL=http://localhost:3333
PORT=3333
```

## Development

### Running Tests

```bash
go test ./...
```

### Database Migrations

The application uses a custom migration system. Migrations are automatically applied on startup.

### Code Structure

The application follows Clean Architecture principles:

- **Domain Layer**: Business entities and rules
- **Application Layer**: Use cases and application services
- **Infrastructure Layer**: External concerns (database, HTTP)
- **Presentation Layer**: HTTP handlers and templates

### Key Components

#### Domain Models

- **Product**: Represents a product with members and categories
- **Launch**: Represents a product launch with states (draft, review, published)
- **User**: User entity with social accounts and sessions

#### Services

- **AuthService**: Handles authentication and session management
- **ProductService**: Manages product operations
- **LaunchService**: Handles launch workflow and moderation
- **UserService**: User management operations

#### Middleware

- **SessionMiddleware**: Session management
- **AdminMiddleware**: Admin-only route protection
- **RateLimitMiddleware**: Rate limiting for API endpoints

## API Endpoints

### Public Routes

- `GET /` - Home page with latest launches
- `GET /products/{slug}` - Product details
- `GET /u/{userID}` - User profile

### Protected Routes

- `GET /new-product` - Create new product form
- `GET /my/products` - User's products
- `POST /api/new-product` - Create product
- `POST /api/new-launch` - Create launch

### Admin Routes

- `GET /admin/moderation/launches` - Launch moderation
- `POST /api/proceed-launch` - Approve launch
- `POST /api/decline-launch` - Decline launch

### Authentication

- `GET /auth/yandex` - Yandex OAuth login
- `GET /auth/yandex/callback` - OAuth callback
- `GET /auth/google` - Google OAuth login
- `GET /auth/google/callback` - OAuth callback
- `GET /api/login` - Login modal
- `GET /api/logout` - Logout

## Deployment

### Production Considerations

1. **Database**: Consider using PostgreSQL for production
2. **HTTPS**: Always use HTTPS in production
3. **Rate Limiting**: Configure appropriate rate limits
4. **Monitoring**: Add health checks and monitoring
5. **Logging**: Configure structured logging

### Docker Deployment

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License.

## Acknowledgments

- Inspired by Product Hunt
- Built with Go and HTMX
- Styled with Tailwind CSS
