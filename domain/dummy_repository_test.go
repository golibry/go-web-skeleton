package domain

import (
	"context"
	"testing"

	"github.com/golibry/go-web-skeleton/infrastructure/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// DummyRepositoryTestSuite is the test suite for DummyRepository
type DummyRepositoryTestSuite struct {
	suite.Suite
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

func (suite *DummyRepositoryTestSuite) TestItIdentifiesNewDummyByZeroId() {
	// Given
	dummy, _ := NewDummy("test")

	// Then - NewDummy creates entity with zero ID
	assert.Equal(suite.T(), 0, dummy.GetId(), "New dummy should have zero ID")
}

func (suite *DummyRepositoryTestSuite) TestItIdentifiesExistingDummyByNonZeroId() {
	// Given
	dummy := ReconstituteDummy(123, "existing")

	// Then - ReconstituteDummy creates entity with provided ID
	assert.Equal(suite.T(), 123, dummy.GetId(), "Reconstituted dummy should have provided ID")
}

func (suite *DummyRepositoryTestSuite) TestItCanAddIdentityToNewDummy() {
	// Given
	dummy, _ := NewDummy("test")
	assert.Equal(suite.T(), 0, dummy.GetId(), "Initially should have zero ID")

	// When
	dummy.AddIdentity(456)

	// Then
	assert.Equal(suite.T(), 456, dummy.GetId(), "Should have the added identity")
}
