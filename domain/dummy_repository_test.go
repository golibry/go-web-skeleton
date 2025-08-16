package domain

import (
	"context"
	"testing"

	"github.com/golibry/go-web-skeleton/infrastructure/registry"
	"github.com/golibry/go-web-skeleton/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// DummyRepositoryTestSuite is the test suite for DummyRepository
type DummyRepositoryTestSuite struct {
	suite.Suite
	bootstrap  *test.Bootstrap
	repository *DummyRepository
}

// SetupSuite initializes the test environment using the bootstrap mechanism
func (suite *DummyRepositoryTestSuite) SetupSuite() {
	// Use the test bootstrap to set up the environment
	suite.bootstrap, _ = test.GetGlobalTestBootstrap()

	// Initialize repository
	suite.repository = NewDummyRepository(*suite.bootstrap.DbService)
}

// TearDownSuite cleans up resources after all tests
func (suite *DummyRepositoryTestSuite) TearDownSuite() {
	suite.bootstrap.TeardownTestEnvironment()
}

// TearDownTest cleans up test data after each test
func (suite *DummyRepositoryTestSuite) TearDownTest() {
	_, err := suite.bootstrap.DbService.Db().Exec("TRUNCATE TABLE dummy")
	if err != nil {
		suite.T().Fatalf("Warning: Failed to clean up test data: %v", err)
	}
}

// TestDummyRepositoryTestSuite runs the test suite
func TestDummyRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(DummyRepositoryTestSuite))
}

func (suite *DummyRepositoryTestSuite) TestItCanCreateNewDummyRepository() {
	// Given
	dbService := registry.DbService{}

	// When
	repository := NewDummyRepository(dbService)

	// Then
	assert.NotNil(suite.T(), repository)
	assert.Equal(suite.T(), dbService, repository.dbService)
}

func (suite *DummyRepositoryTestSuite) TestItCannotSaveNilDummy() {
	// Given
	ctx := context.Background()
	dbService := registry.DbService{}
	repository := NewDummyRepository(dbService)

	// When
	err := repository.Save(ctx, nil)

	// Then
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "dummy cannot be nil")
}

func (suite *DummyRepositoryTestSuite) TestItCannotSaveWhenDatabaseConnectionIsNil() {
	// Given
	ctx := context.Background()
	dummy, _ := NewDummy("test")
	dbService := registry.DbService{} // empty service with nil db
	repository := NewDummyRepository(dbService)

	// When
	err := repository.Save(ctx, dummy)

	// Then
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "database connection is nil")
}

func (suite *DummyRepositoryTestSuite) TestItCanPerformHappyUpsertFlow() {
	// Given
	ctx := context.Background()

	// When - Create new dummy (INSERT operation)
	dummy, err := NewDummy("Initial Name")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, dummy.GetId()) // New entity should have ID = 0

	// Save the new dummy (should perform INSERT)
	err = suite.repository.Save(ctx, dummy)

	// Then - Verify INSERT worked
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), dummy.GetId(), 0) // Should have auto-generated ID
	insertedId := dummy.GetId()
	assert.Equal(suite.T(), "Initial Name", dummy.GetName())

	// When - Update the existing dummy (UPDATE operation)
	// Create a new dummy with the same ID but different name
	updatedDummy := ReconstituteDummy(insertedId, "Updated Name")
	assert.Equal(suite.T(), insertedId, updatedDummy.GetId()) // Should have the existing ID

	// Save the updated dummy (should perform UPDATE)
	err = suite.repository.Save(ctx, updatedDummy)

	// Then - Verify UPDATE worked
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), insertedId, updatedDummy.GetId()) // ID should remain the same
	assert.Equal(suite.T(), "Updated Name", updatedDummy.GetName())

	// Additional verification - Query the database directly to confirm the update
	var dbName string
	err = suite.bootstrap.DbService.Db().QueryRowContext(
		ctx,
		"SELECT name FROM dummy WHERE id = ?",
		insertedId,
	).Scan(&dbName)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Name", dbName)
}
