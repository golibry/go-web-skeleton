package registry

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golibry/go-common-domain/domain"
	httplib "github.com/golibry/go-http/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ResponseBuilderServiceTestSuite is the test suite for ResponseBuilder
type ResponseBuilderServiceTestSuite struct {
	suite.Suite
	service *ResponseBuilder
	logger  *LoggerService
}

// SetupTest runs before each test
func (suite *ResponseBuilderServiceTestSuite) SetupTest() {
	// Create a simple logger service without config dependency for testing
	mockLogger := slog.New(
		slog.NewTextHandler(
			os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			},
		),
	)

	suite.logger = &LoggerService{
		logger:    mockLogger,
		logWriter: &stdoutWriter{},
	}

	suite.service = NewResponseBuilderService(suite.logger)
	suite.Require().NotNil(suite.service, "ResponseBuilder should not be nil")
}

// TearDownTest runs after each test
func (suite *ResponseBuilderServiceTestSuite) TearDownTest() {
	// No cleanup needed for mock logger
}

// Test service initialization
func (suite *ResponseBuilderServiceTestSuite) TestItCanCreateNewResponseBuilderService() {
	assert.NotNil(suite.T(), suite.service)
	assert.NotNil(suite.T(), suite.service.logger)
	assert.NotNil(suite.T(), suite.service.ctx)
	assert.Equal(suite.T(), context.Background(), suite.service.ctx)
}

// Test NewBuilder functionality
func (suite *ResponseBuilderServiceTestSuite) TestItCanCreateNewBuilder() {
	recorder := httptest.NewRecorder()
	builder := suite.service.NewBuilder(recorder)

	assert.NotNil(suite.T(), builder)
	assert.IsType(suite.T(), &httplib.ResponseBuilder{}, builder)
}

// Test NewBuilder can create specialized builders
func (suite *ResponseBuilderServiceTestSuite) TestItCanCreateSpecializedBuildersFromNewBuilder() {
	recorder := httptest.NewRecorder()
	builder := suite.service.NewBuilder(recorder)

	// Test JSON builder creation
	jsonBuilder := builder.JSON()
	assert.NotNil(suite.T(), jsonBuilder)
	assert.IsType(suite.T(), &httplib.JSONResponseBuilder{}, jsonBuilder)

	// Test Text builder creation
	textBuilder := builder.Text()
	assert.NotNil(suite.T(), textBuilder)
	assert.IsType(suite.T(), &httplib.TextResponseBuilder{}, textBuilder)

	// Test HTML builder creation
	htmlBuilder := builder.HTML()
	assert.NotNil(suite.T(), htmlBuilder)
	assert.IsType(suite.T(), &httplib.HTMLResponseBuilder{}, htmlBuilder)

	// Test Error builder creation
	errorBuilder := builder.Error()
	assert.NotNil(suite.T(), errorBuilder)
	assert.IsType(suite.T(), &httplib.ErrorResponseBuilder{}, errorBuilder)
}

// Test NewBuilder with status and headers
func (suite *ResponseBuilderServiceTestSuite) TestItCanCreateBuilderWithStatusAndHeaders() {
	recorder := httptest.NewRecorder()
	builder := suite.service.NewBuilder(recorder)

	// Test builder chaining
	chainedBuilder := builder.Status(http.StatusCreated).Header("X-Test", "test-value")
	assert.NotNil(suite.T(), chainedBuilder)
	assert.Equal(suite.T(), builder, chainedBuilder) // Should return same instance for chaining

	// Test that status and headers are applied when using a specialized builder
	jsonBuilder := chainedBuilder.JSON()
	data := map[string]string{"message": "test"}
	err := jsonBuilder.Data(data).Send()

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, recorder.Code)
	assert.Equal(suite.T(), "test-value", recorder.Header().Get("X-Test"))
	assert.Equal(suite.T(), "application/json", recorder.Header().Get("Content-Type"))
}

// Test NewErrorBuilder functionality
func (suite *ResponseBuilderServiceTestSuite) TestItCanCreateNewErrorBuilder() {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/test", nil)
	
	errorBuilder := suite.service.NewErrorBuilder(recorder, request)

	assert.NotNil(suite.T(), errorBuilder)
	assert.IsType(suite.T(), &httplib.ErrorResponseBuilder{}, errorBuilder)
}

// Test NewErrorBuilder with error handling
func (suite *ResponseBuilderServiceTestSuite) TestItCanHandleErrorsWithNewErrorBuilder() {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/test", nil)
	
	errorBuilder := suite.service.NewErrorBuilder(recorder, request)
	testErr := errors.New("test error")
	
	err := errorBuilder.WithError(testErr).Send()
	
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusInternalServerError, recorder.Code)
	assert.Contains(suite.T(), recorder.Body.String(), "test error")
}

// Test NewErrorBuilder with domain error (should map to bad request)
func (suite *ResponseBuilderServiceTestSuite) TestItCanHandleDomainErrorsWithNewErrorBuilder() {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/test", nil)
	
	errorBuilder := suite.service.NewErrorBuilder(recorder, request)
	// Create a domain error that should map to BadRequest based on setupErrorCategories
	domainErr := domain.NewError("validation failed")
	
	err := errorBuilder.WithError(domainErr).Send()
	
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusBadRequest, recorder.Code)
	assert.Contains(suite.T(), recorder.Body.String(), "validation failed")
}

// Test setupErrorCategories function
func (suite *ResponseBuilderServiceTestSuite) TestItCanSetupErrorCategories() {
	categories := setupErrorCategories()
	
	assert.NotNil(suite.T(), categories)
	assert.Len(suite.T(), categories, 1)
	
	badRequestCategory := categories[0]
	assert.NotNil(suite.T(), badRequestCategory)
	
	// Test that domain errors match the bad request category
	domainErr := domain.NewError("test domain error")
	matches := badRequestCategory.Matches(domainErr)
	assert.True(suite.T(), matches, "Domain error should match bad request category")
}

// Run the test suite
func TestResponseBuilderServiceSuite(t *testing.T) {
	suite.Run(t, new(ResponseBuilderServiceTestSuite))
}
