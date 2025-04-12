# Contributing

Thank you for your interest in contributing to the project! This document provides guidelines and instructions for contributing.

## Code of Conduct

Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md) in all your interactions with the project.

## Development Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/cool-new-feature`)
3. Make your changes
4. Run tests and linters:

    ```bash
    make test
    make lint
    ```

5. Commit your changes (`git commit -m 'Add a cool new feature'`)
6. Push to the branch (`git push origin feature/cool-new-feature`)
7. Open a Pull Request

## Pull Request Process

1. Ensure your PR description clearly describes the problem and solution
2. Update the README.md if necessary
3. The PR must pass all CI checks
4. A maintainer will review your PR and merge it once approved

## Development Setup

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

4. Start the development server:

    ```bash
    make dev
    ```

   Or run directly:

    ```bash
    go run main.go server
    ```

   The server will start on port 4000 by default. You can change the port using the `-p` flag:

    ```bash
    go run main.go server -p 8080
    ```

## Code Style

- Follow the [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- Use `gofmt` for Go code formatting
- Follow the project's ESLint and Prettier configurations for frontend code
- Write tests for new features and bug fixes

## Testing

- Run the test suite before submitting a PR:

    ```bash
    make test
    ```

- Ensure test coverage doesn't decrease
- Add tests for new features

## Security

If you discover a security vulnerability, please use GitHub's private vulnerability reporting feature instead of opening a public issue. This helps protect our users while we work on a fix.

## Questions?

Feel free to open an issue if you have any questions about contributing to the project.
