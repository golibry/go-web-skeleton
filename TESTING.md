# Testing Guide

This guide explains how to set up and use the test bootstrap mechanism in the Go Web Skeleton project.

## Test Bootstrap Overview

The project provides a test bootstrap mechanism to standardize test environment setup across all test packages. This ensures consistent configuration, database connections, and cleanup procedures.

## Bootstrap Components

### TestBootstrap (`infrastructure/registry/test_bootstrap.go`)

The `TestBootstrap` struct provides centralized test environment setup:

```go
type TestBootstrap struct {
    ConfigService *ConfigService
    DbService     *DbService
    LoggerService *LoggerService
}
```

### Key Functions

#### `SetupTestEnvironment()`
Initializes the test environment with proper configuration:
- Sets `APP_ENV=test` if not already set
- Loads test-specific environment variables from `.env.test.local`
- Initializes all core services (Config, Database, Logger)

#### `TeardownTestEnvironment()`
Cleans up resources after tests complete:
- Closes database connections
- Performs necessary cleanup

#### `CleanupTestData()`
Removes test data between test runs:
- Cleans up all test tables
- Ensures test isolation

## Usage Patterns

### Pattern 1: TestMain Bootstrap (Recommended)

Use this pattern for package-level test setup:

```go
package domain

import (
    "testing"
    "github.com/golibry/go-web-skeleton/infrastructure/registry"
)

func TestMain(m *testing.M) {
    registry.TestMainBootstrap(m)
}
```

### Pattern 2: Test Suite Bootstrap

Use this pattern for testify/suite integration:

```go
package domain

import (
    "testing"
    "github.com/golibry/go-web-skeleton/infrastructure/registry"
    "github.com/stretchr/testify/suite"
)

type MyTestSuite struct {
    suite.Suite
    bootstrap  *registry.TestBootstrap
    // your test dependencies
}

func (suite *MyTestSuite) SetupSuite() {
    var err error
    suite.bootstrap, err = registry.SetupTestEnvironment()
    if err != nil {
        suite.T().Fatalf("Failed to setup test environment: %v", err)
    }
    
    // Initialize your test dependencies using bootstrap services
    // suite.repository = NewRepository(*suite.bootstrap.DbService)
}

func (suite *MyTestSuite) TearDownSuite() {
    if suite.bootstrap != nil {
        suite.bootstrap.TeardownTestEnvironment()
    }
}

func (suite *MyTestSuite) TearDownTest() {
    if suite.bootstrap != nil {
        if err := suite.bootstrap.CleanupTestData(); err != nil {
            suite.T().Logf("Warning: Failed to clean up test data: %v", err)
        }
    }
}

func TestMyTestSuite(t *testing.T) {
    suite.Run(t, new(MyTestSuite))
}
```

## Environment Configuration

### Test Environment Files

The bootstrap mechanism loads environment variables in the following priority order:

1. `.env.test.local` (highest priority - for local test configuration)
2. `.env.local` (skipped in test environment)  
3. `.env.test` (shared test configuration)
4. `.env` (lowest priority - base configuration)

### Required Test Configuration

Your `.env.test.local` file must contain:

```bash
# Test environment configuration
APP_BASE_DIR=/path/to/your/project
APP_ENV=test
APP_LOG_LEVEL=info
APP_LOG_PATH=stdout

# Database Configuration for Test Environment
DB_DSN=user:password@tcp(localhost:3306)/test_database
DB_MAX_IDLE_CONNECTIONS=2
DB_MAX_OPEN_CONNECTIONS=10
DB_CONNECTION_MAX_IDLE_TIME=3m
DB_CONNECTION_MAX_LIFETIME=3m
DB_MIGRATIONS_DIR_PATH=/path/to/your/project/migrations/versions

# HTTP Server Configuration
HTTP_BIND_ADDRESS=127.0.0.1
HTTP_BIND_PORT=8081
HTTP_MAX_HEADER_BYTES=16384
HTTP_REQUEST_TIMEOUT=15s
```

## Best Practices

### 1. Use Consistent Setup
Always use the bootstrap mechanism for integration tests that require database or configuration access.

### 2. Test Isolation
Use `CleanupTestData()` in `TearDownTest()` methods to ensure test isolation:

```go
func (suite *MyTestSuite) TearDownTest() {
    if suite.bootstrap != nil {
        suite.bootstrap.CleanupTestData()
    }
}
```

### 3. Error Handling
Always check for errors during setup and fail fast:

```go
suite.bootstrap, err = registry.SetupTestEnvironment()
if err != nil {
    suite.T().Fatalf("Failed to setup test environment: %v", err)
}
```

### 4. Resource Management
Always call `TeardownTestEnvironment()` to prevent resource leaks:

```go
func (suite *MyTestSuite) TearDownSuite() {
    if suite.bootstrap != nil {
        suite.bootstrap.TeardownTestEnvironment()
    }
}
```

## Example Implementation

Here's a complete example of a repository test using the bootstrap mechanism:

```go
package domain

import (
    "context"
    "testing"

    "github.com/golibry/go-web-skeleton/infrastructure/registry"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
)

type DummyRepositoryTestSuite struct {
    suite.Suite
    bootstrap  *registry.TestBootstrap
    repository *DummyRepository
}

func (suite *DummyRepositoryTestSuite) SetupSuite() {
    var err error
    suite.bootstrap, err = registry.SetupTestEnvironment()
    if err != nil {
        suite.T().Fatalf("Failed to setup test environment: %v", err)
    }
    suite.repository = NewDummyRepository(*suite.bootstrap.DbService)
}

func (suite *DummyRepositoryTestSuite) TearDownSuite() {
    if suite.bootstrap != nil {
        suite.bootstrap.TeardownTestEnvironment()
    }
}

func (suite *DummyRepositoryTestSuite) TearDownTest() {
    if suite.bootstrap != nil {
        if err := suite.bootstrap.CleanupTestData(); err != nil {
            suite.T().Logf("Warning: Failed to clean up test data: %v", err)
        }
    }
}

func (suite *DummyRepositoryTestSuite) TestItCanSaveDummy() {
    // Given
    ctx := context.Background()
    dummy, _ := NewDummy("Test Name")

    // When
    err := suite.repository.Save(ctx, dummy)

    // Then
    assert.NoError(suite.T(), err)
    assert.Greater(suite.T(), dummy.GetId(), 0)
}

func TestDummyRepositoryTestSuite(t *testing.T) {
    suite.Run(t, new(DummyRepositoryTestSuite))
}
```

## Troubleshooting

### Configuration Issues
If you encounter configuration validation errors:
1. Verify your `.env.test.local` file exists and contains all required variables
2. Check that file paths are absolute and correct for your system
3. Ensure database connection details are valid

### Database Connection Issues
If database tests fail:
1. Verify your test database exists and is accessible
2. Check database credentials in `.env.test.local`
3. Ensure test database schema is up to date

### Environment Variable Issues
If environment variables aren't loading:
1. Check file paths in your configuration
2. Verify working directory matches your project root
3. Ensure `.env.test.local` file permissions are readable

## Integration with CI/CD

For continuous integration, create a `.env.test` file (without `.local`) with CI-appropriate values:

```bash
# CI Test Configuration
APP_ENV=test
DB_DSN=testuser:testpass@tcp(test-db:3306)/test_db
DB_MIGRATIONS_DIR_PATH=/github/workspace/migrations/versions
```

This approach allows local developers to use `.env.test.local` while CI systems use `.env.test`.