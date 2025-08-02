# nullmail

A development SMTP server that captures emails without actually sending them, with a modern web interface for viewing received messages.

## Features

- **SMTP Server**: Accepts emails on port 2525 (or configurable port)
- **Web Interface**: Modern Next.js client for viewing captured emails
- **Redis Storage**: Persistent email storage using Redis
- **TLS Support**: Supports encrypted connections
- **Development Ready**: Easy setup for local development and testing

## Quick Start

### Prerequisites

- Go 1.21.3 or higher
- Node.js and pnpm
- Docker and Docker Compose (for Redis)

### Setup

1. **Start Redis services**:
   ```bash
   make docker-up
   ```

2. **Run the SMTP server**:
   ```bash
   make run
   ```

3. **Start the web client** (in another terminal):
   ```bash
   make run-client
   ```

4. **Access the web interface** at `http://localhost:3000`

### Testing Email Reception

Send a test email to the SMTP server:

```bash
make test-smtp
```

Or manually with `swaks`:
```bash
swaks --to user@nullmail.local --from sender@example.com --server localhost:2525 --body "Test email"
```

## Development

### Available Commands

**Go/SMTP Server:**
- `make run` - Start SMTP server in development mode
- `make build` - Build the binary
- `make test` - Run tests
- `make fmt` - Format Go code
- `make vet` - Run go vet
- `make check` - Run formatting and vetting

**Client:**
- `make run-client` - Start Next.js development server

**Docker:**
- `make docker-up` - Start all services (Redis, etc.)
- `make docker-down` - Stop all services
- `make docker-build` - Build Docker image

**Redis:**
- `make redis-logs` - View Redis logs
- `make redis-reset` - Clear all stored emails

### Project Structure

```
nullmail/
├── cmd/nullmail/          # Main application entry point
├── internal/              # Internal Go packages
│   ├── smtp/             # SMTP server implementation
│   ├── email/            # Email parsing and validation
│   └── redis/            # Redis client
├── client/               # Next.js web interface
│   ├── app/              # Next.js app router
│   ├── components/       # React components
│   └── lib/              # Client utilities
└── infra/                # Infrastructure configs
```

## Configuration

### Environment Variables

**SMTP Server:**
- `DEBUG=true` - Enable debug logging
- `ENV=production` - Set production mode
- `REDIS_URL` - Redis connection string (default: localhost:6379)
- `REDIS_PASSWORD` - Redis password (default: dev123)

**Client:**
- Standard Next.js environment variables

### Ports

- SMTP Server: 2525 (configurable via command line argument)
- Web Client: 3000 (Next.js default)
- Redis: 6379

## API

The web client connects to the SMTP server's stored emails via:

- `GET /api/emails/[address]` - Retrieve emails for a specific address

## License

This project is for development and testing purposes.