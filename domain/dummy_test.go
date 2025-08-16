package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// DummyTestSuite is the test suite for Dummy entity
type DummyTestSuite struct {
	suite.Suite
}

// TestDummyTestSuite runs the test suite
func TestDummyTestSuite(t *testing.T) {
	suite.Run(t, new(DummyTestSuite))
}

func (suite *DummyTestSuite) TestItIdentifiesNewDummyByZeroId() {
	// Given
	dummy, _ := NewDummy("test")

	// Then - NewDummy creates an entity with zero ID
	assert.Equal(suite.T(), 0, dummy.GetId(), "New dummy should have zero ID")
}

func (suite *DummyTestSuite) TestItIdentifiesExistingDummyByNonZeroId() {
	// Given
	dummy := ReconstituteDummy(123, "existing")

	// Then - ReconstituteDummy creates an entity with provided ID
	assert.Equal(suite.T(), 123, dummy.GetId(), "Reconstituted dummy should have provided ID")
}

func (suite *DummyTestSuite) TestItCanAddIdentityToNewDummy() {
	// Given
	dummy, _ := NewDummy("test")
	assert.Equal(suite.T(), 0, dummy.GetId(), "Initially should have zero ID")

	// When
	dummy.AddIdentity(456)

	// Then
	assert.Equal(suite.T(), 456, dummy.GetId(), "Should have the added identity")
}
