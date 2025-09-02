# Airlock

A Go-based authentication service that provides passwordless email authentication using AWS services (DynamoDB, SES). Serves a web frontend and provides REST API endpoints for user authentication via JWT tokens.

## Quick Start

### Prerequisites
- Go 1.24.4+
- AWS credentials configured
- Make

### Development
```bash
# Build for local development
make dev

# Run the service
./bin/airlock-darwin-arm64
```

### Production
```bash
# Build for production
make prod

# Deploy to remote server
make deploy
```

## API Endpoints

### Authentication
- `POST /api/auth/email` - Request email authentication
- `GET /api/auth/email/verify` - Verify email token (returns JWT)

### Authentication Flow
1. User requests authentication with email
2. System generates token, sends verification email
3. User clicks email link to verify token
4. System returns JWT bearer token (30-day expiry)

## Configuration

Create `.env` file with required variables:
```env
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
SERVICE_URL=https://your-domain.com
EMAIL_AUTH_DEBOUNCE=180
EMAIL_AUTH_EXPIRY=900
PORT=8080
```

## Database

Uses DynamoDB table "Luna4Users" with:
- Primary key: `id` (string)
- Global secondary index: `email-index`

## Build Commands

- `make dev` - Build for Darwin ARM64
- `make prod` - Build for Linux AMD64
- `make all` - Build both targets
- `make clean` - Remove build artifacts
- `make deploy` - Full deployment pipeline

## Architecture

```
internal/
├── handler/     # HTTP request handlers
├── service/     # Business logic (User, Email, DynamoDB)
└── model/       # Data models

util/           # Authentication utilities
web/            # Static frontend files (embedded)
assets/         # Email templates
```

## License

Private project.