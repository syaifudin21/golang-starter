# Go Exam Project

This is a Go project that demonstrates a simple API with user authentication, authorization, and internationalization.

## Features

- **User Authentication**: Login, logout, and token refresh.
- **JWT Middleware**: Secures endpoints using JSON Web Tokens.
- **Casbin Middleware**: Provides Role-Based Access Control (RBAC).
- **Database Migration**: Uses `golang-migrate` to manage database schema changes.
- **Validation**: Uses `go-playground/validator` for request validation.
- **Internationalization (i18n)**: Supports English and Indonesian languages for validation messages.
- **Configuration**: Uses `.env` files for easy configuration.

## Installed Packages

### Direct Dependencies
- github.com/casbin/casbin/v2
- github.com/go-sql-driver/mysql
- github.com/golang-jwt/jwt/v5
- github.com/golang-migrate/migrate/v4
- github.com/google/uuid
- github.com/joho/godotenv
- github.com/labstack/echo/v4
- golang.org/x/crypto
- github.com/go-playground/validator/v10
- github.com/nicksnyder/go-i18n/v2

### Indirect Dependencies
- filippo.io/edwards25519
- github.com/bmatcuk/doublestar/v4
- github.com/casbin/govaluate
- github.com/gabriel-vasile/mimetype
- github.com/go-playground/locales
- github.com/go-playground/universal-translator
- github.com/hashicorp/errwrap
- github.com/hashicorp/go-multierror
- github.com/labstack/gommon
- github.com/leodido/go-urn
- github.com/mattn/go-colorable
- github.com/mattn/go-isatty
- github.com/valyala/bytebufferpool
- github.com/valyala/fasttemplate
- golang.org/x/net
- golang.org/x/sys
- golang.org/x/text

## Getting Started

### Prerequisites

- Go 1.24.5 or higher
- MySQL

### Installation

1.  Clone the repository:
    ```bash
    git clone <repository-url>
    ```
2.  Install dependencies:
    ```bash
    go mod tidy
    ```
3.  Create a `.env` file and configure your database connection. See `.env.example` for reference.
4.  Run database migrations:
    ```bash
    go run . migrate up
    ```
5.  Start the API server:
    ```bash
    go run . api
    ```

The server will start on the port specified in your `.env` file (default is 8080).
