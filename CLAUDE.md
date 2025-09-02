# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Airlock is a Go-based authentication service that provides passwordless email authentication using AWS services (DynamoDB, SES). It serves a web frontend and provides REST API endpoints for user authentication via JWT tokens.

## Build and Development Commands

### Building the Application
- `make dev` - Build for local development (Darwin ARM64)
- `make prod` - Build for production deployment (Linux AMD64) 
- `make all` - Build both development and production binaries
- `make clean` - Remove build artifacts from bin/ directory

### Deployment
- `make deploy` - Complete deployment pipeline (build, copy files, restart service on remote server)
- Individual deployment scripts in `scripts/`:
  - `update-remote-bin.sh` - Copy production binary to remote server
  - `update-remote-service-conf.sh` - Copy systemd service configuration  
  - `remote-service-up.sh` - Start the remote service
  - `remote-service-down.sh` - Stop the remote service
  - `ssh-remote.sh` - SSH into the remote server

### Environment Configuration
- `.env` - Development environment variables
- `.deploy.env` - Deployment configuration (SSH keys, remote server details)

## Code Architecture

### Application Structure
- **main.go** - Entry point, sets up Gin router, embedded static files, API routes
- **internal/handler/** - HTTP request handlers for API endpoints
- **internal/service/** - Business logic layer (User, Email, DynamoDB services)
- **internal/model/** - Data models and structs
- **util/** - Utility functions for authentication, token generation
- **web/** - Static frontend files (HTML, CSS, JS) embedded at build time
- **assets/templates/** - Email templates for authentication emails

### Key Services
- **UserService** - Handles user data operations with DynamoDB, manages email authentication flow
- **EmailService** - Sends authentication emails via AWS SES using HTML templates
- **DynamoDBService** - Low-level database operations, connection management

### Authentication Flow
1. POST `/api/auth/email` - Request email authentication (generates token, sends email)
2. GET `/api/auth/email/verify` - Verify email token (returns JWT bearer token)
3. Email tokens are hashed with SHA256, stored in DynamoDB with debounce/expiry logic
4. JWT tokens are signed with HS256, expire in 30 days

### Database Schema
DynamoDB table "Luna4Users" with:
- Primary key: `id` (string)
- Global secondary index: `email-index` on email field
- User model includes email authentication state, preferences, timestamps

### Environment Variables
Key variables (see .env files for complete list):
- `AWS_REGION`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` - AWS credentials
- `EMAIL_AUTH_DEBOUNCE` - Seconds between email requests (default 180)
- `EMAIL_AUTH_EXPIRY` - Token expiry in seconds (default 900)
- `SERVICE_URL` - Domain for email verification links
- `PORT` - Server port (default varies by environment)

### Deployment Architecture
- Production builds create Linux AMD64 binary at `bin/airlock-linux-amd64`
- Systemd service configuration in `deployments/airlock.service`
- Remote deployment to EC2 instance via SSH with automated service management
- All deployment scripts use relative paths from script directory for portability

## File Path Conventions

When working with files in this codebase:
- Build artifacts go in `bin/`
- Environment files are in project root
- Deployment configs are in `deployments/`
- All scripts reference files relative to `$PROJECT_ROOT` for portability
- Static assets are embedded at build time, templates loaded from `assets/templates/`