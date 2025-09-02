.PHONY: dev prod clean deploy run logs

# Build for local development (Apple Silicon)
dev:
	@mkdir -p bin
	GOOS=darwin GOARCH=arm64 go build -o bin/airlock-darwin-arm64 main.go

# Build for AWS AMI2 (Linux x86_64)
prod:
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -o bin/airlock-linux-amd64 main.go

# Run development server
run:
	@bash -c "source .env && go run ."

# Clean build artifacts
clean:
	rm -rf bin/

# Download logs from remote server
logs:
	@./scripts/download-remote-logs.sh

# Deploy to remote server
deploy:
	@./scripts/deploy-remote.sh

# Build both targets
all: dev prod