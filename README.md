# Go Web Starter

[![CI Status](https://github.com/Dobefu/go-web-starter/actions/workflows/ci.yml/badge.svg)](https://github.com/Dobefu/go-web-starter/actions/workflows/ci.yml)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=Dobefu_go-web-starter&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=Dobefu_go-web-starter)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=Dobefu_go-web-starter&metric=coverage)](https://sonarcloud.io/summary/new_code?id=Dobefu_go-web-starter)
[![Go Report Card](https://goreportcard.com/badge/github.com/Dobefu/go-web-starter)](https://goreportcard.com/report/github.com/Dobefu/go-web-starter)

> [!WARNING]
> This repository is still a work-in-progress

A modern, production-ready Go web application starter template with best practices and common features pre-configured. This template includes both backend (Go) and frontend (TypeScript) components.

## Features

- 🚀 Fast and efficient web server using [Gin](https://github.com/gin-gonic/gin)
- 📦 Clean project structure following Go best practices
- 🔧 Live reloading for development
- 🧪 Built-in testing setup
- 📊 Code quality tools (SonarQube, ESLint, Prettier)
- 🔄 CI/CD pipeline ready
- 🛡️ Security best practices
- 💻 Modern frontend development with TypeScript and Bun
- 🎨 Consistent code formatting with Prettier
- 📝 Type safety with TypeScript

## Prerequisites

- Go 1.24 or higher
- Bun (for frontend development)
- Air (optional, for live reloading)
- Make (optional, for using Makefile commands)
- Docker (optional, for containerization)

## Getting Started

### Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/Dobefu/go-web-starter.git
   cd go-web-starter
   ```

2. Install Go dependencies:

   ```bash
   go mod download
   ```

3. Install frontend dependencies:

   ```bash
   bun install
   ```

### Development

1. Start the development server with live reloading:

   ```bash
   make dev
   ```

   Or run directly:

   ```bash
   go run main.go server
   ```

2. The server will start on port 4000 by default. You can change the port using the `-p` flag:

   ```bash
   go run main.go server -p 8080
   ```

### Testing

Run the test suite:

```bash
make test
```

### Building

Build the application:

```bash
make build
```

This will build both the frontend and backend components.

## Development Tools

### Air (For live Reloading)

The project uses [Air](https://github.com/cosmtrek/air) for live reloading during development. Configuration can be found in `.air.toml`.

### Frontend Development

The frontend uses:

- TypeScript for type safety
- Bun for package management and bundling
- ESLint for code linting
- Prettier for code formatting

### Make Commands

- `make dev`: Start development server with live reloading
- `make build`: Build the application
- `make test`: Run tests
- `make lint`: Run linters
- `make bench`: Run benchmarks
- `make clean`: Clean build artifacts

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/cool-new-feature`)
3. Commit your changes (`git commit -m 'Add a cool new feature'`)
4. Push to the branch (`git push origin feature/cool-new-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For support, please open an issue in the GitHub repository.
