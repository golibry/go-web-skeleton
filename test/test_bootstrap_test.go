package test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// TestBootstrapTestSuite tests the bootstrap functionality
type TestBootstrapTestSuite struct {
	suite.Suite
}

// TestTestBootstrapTestSuite runs the test suite
func TestTestBootstrapTestSuite(t *testing.T) {
	suite.Run(t, new(TestBootstrapTestSuite))
}

func (suite *TestBootstrapTestSuite) TestItCanSetupTestEnvironment() {
	// When
	bootstrap, err := SetupTestEnvironment()

	// Then
	assert.NoError(suite.T(), err, "SetupTestEnvironment should not return an error")
	assert.NotNil(suite.T(), bootstrap, "Bootstrap should not be nil")
	assert.NotNil(suite.T(), bootstrap.ConfigService, "ConfigService should be initialized")
	assert.NotNil(suite.T(), bootstrap.LoggerService, "LoggerService should be initialized")
	assert.NotNil(suite.T(), bootstrap.DbService, "DbService should be initialized")

	// Verify environment is set to test
	assert.Equal(suite.T(), "test", os.Getenv("APP_ENV"), "APP_ENV should be set to test")

	// Cleanup
	bootstrap.TeardownTestEnvironment()
}

func (suite *TestBootstrapTestSuite) TestItCanCleanupTestData() {
	// Given
	bootstrap, err := SetupTestEnvironment()
	assert.NoError(suite.T(), err)
	defer bootstrap.TeardownTestEnvironment()

	// When
	err = bootstrap.RemoveTestDb()

	// Then
	assert.NoError(suite.T(), err, "RemoveTestDb should not return an error")
}

func (suite *TestBootstrapTestSuite) TestItSetsTestEnvironmentVariable() {
	// Given - Clear APP_ENV to test the setting behavior
	originalEnv := os.Getenv("APP_ENV")
	os.Unsetenv("APP_ENV")
	defer func() {
		if originalEnv != "" {
			os.Setenv("APP_ENV", originalEnv)
		}
	}()

	// When
	bootstrap, err := SetupTestEnvironment()

	// Then
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test", os.Getenv("APP_ENV"))

	// Cleanup
	bootstrap.TeardownTestEnvironment()
}
