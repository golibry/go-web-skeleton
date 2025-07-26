## Overview
The project's purpose is to provide a web application skeleton to help developers jump in 
quickly and start building business features without having to build all the necessary 
boilerplate for a web application.  

# Development Workflow

## Step-by-Step Development Process
1. **Understand**: Analyze the requirement and identify affected components
2. **Plan**: Design the solution considering middleware patterns and error handling
3. **Implement**: Write code following established patterns (see the Code Patterns section)
4. **Test**: Create comprehensive tests using testify suite
5. **Validate**: Ensure integration with the existing middleware chain works correctly

## Core Principles
- **Iterative Approach**: Develop incrementally with frequent validation
- **Domain Alignment**: Use consistent HTTP/middleware terminology throughout
- **Evidence-Based**: All decisions must be testable and measurable
- **Context Awareness**: Maintain understanding of the entire middleware chain
- **Structured Execution**: Always plan before implementing
- **Maintenance costs**: Add enough code that justifies the return on investment (more lines =
  more maintenance costs)

# Development Principles
DRY: Abstract common functionality, remove duplication
KISS: Prefer simplicity to complexity in all design decisions
YAGNI: Implement only current requirements, avoid speculative features
Separation of Concerns: Divide program functionality into distinct sections
Loose Coupling: Minimize dependencies between components
High Cohesion: Related functionality should be grouped together logically

# Security
Audit and fix for OWASP top security vulnerabilities

### Environment Setup

The application uses environment-based configuration with the following priority order:
1. `.env.{env}.local` (highest priority)
2. `.env.local` (not loaded in test environment)
3. `.env.{env}`
4. `.env` (lowest priority)

## Testing Practices

### Test Organization
- Use table-driven tests for multiple test cases
- Use testify/suite for all tests
- Group related tests in test suites
- Use descriptive test names that explain the scenario
- Test package behavior and follow the test name convention like TestItCanDoSomething
- Maximize the return on investment to ease maintenance
- Do not go more than once through a tested logic path  

### Error Testing
- Test both success and failure scenarios
- Use `errors.Is()` and `errors.As()` in tests
- Test error wrapping and unwrapping behavior
- Verify error messages are meaningful

### Test Coverage
- Test all public methods and functions
- Test edge cases and boundary conditions
- Test error conditions and invalid inputs
- Test JSON marshaling/unmarshaling if applicable

## Code Documentation

### Function Documentation
- Document all exported functions and methods
- Start with the function name
- Explain what the function does, not how it does it
- Document parameters and return values when not obvious

### Comment Style
- Use complete sentences in comments
- Keep comments concise but informative
- Update comments when code changes
- Avoid obvious comments that just repeat the code

## General Go Best Practices

### Code Organization
- Keep functions small and focused on a single responsibility
- Use early returns to reduce nesting
- Group related functionality together
- Separate concerns are clear
- Do not use both value and pointer receivers in struct methods. Use one or the other

### Concurrency
- Follow Go's concurrency patterns: "Don't communicate by sharing memory; share memory by communicating"
- Use channels for communication between goroutines
- Use sync package primitives when appropriate
- Always handle the goroutine lifecycle properly

### Dependencies
- Keep dependencies minimal and well-justified
- Prefer a standard library when possible
- Use semantic versioning for your modules
- Regularly update dependencies for security patches

### Code Style
- Use `gofmt` to format code consistently
- Use `golint` and `go vet` for code quality checks
- Follow the Go Code Review Comments guidelines
- Use meaningful variable names, even if they're longer

## Development Information

### Architecture

The project follows **Clean Architecture** principles with a clear separation of concerns:

```
├── domain/           # Business entities and value objects
├── infrastructure/   # External concerns (database, registry)
├── presentation/     # HTTP handlers, routes, UI
├── config/          # Configuration management
└── migrations/      # Database schema migrations
```

### Code Organization

**Domain Layer (`domain/`):**
- Contains business entities and value objects
- No external dependencies
- Pure business logic with validation
- Example: `Email`, `FullName`, `Pet`, `Owner`

**Infrastructure Layer (`infrastructure/`):**
- Database connections and repositories
- Dependency injection container
- External service integrations
- Registry pattern for service management

**Presentation Layer (`presentation/`):**
- HTTP server setup and routing
- Request/response handling
- UI templates and static files
- Located in `presentation/http/`

**Configuration (`config/`):**
- Environment-based configuration
- Validation using `github.com/go-playground/validator/v10`
- Structured configuration with separate concerns (HTTP, DB)

### Dependency Injection

The project uses a custom dependency injection container (`infrastructure/registry/container.go`)  

**Service initialization order:**
1. ConfigService (loads and validates configuration)
2. DbService (establishes database connection)
3. LoggerService (structured JSON logging)
4. Domain-specific services (PaymentMethodService, etc.)

### Logging

- Uses structured JSON logging via `log/slog`
- Configurable log levels: `debug`, `info`, `warn`, `error`
- Default output to stdout
- Logger available through container: `container.Logger()`