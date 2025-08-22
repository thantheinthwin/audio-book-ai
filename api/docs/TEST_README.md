# API Tests

This directory contains comprehensive tests for the audio book API handlers.

## Test Structure

The tests follow Go's standard testing conventions as described in the [Go testing tutorial](https://go.dev/doc/tutorial/add-a-test).

### Files

- `audiobooks_test.go` - Tests for the audiobooks handler functions

### Test Coverage

The tests cover the following audiobooks handler functions:

1. **GetAudioBooks** - Tests fetching audiobooks for authenticated users
2. **GetAudioBook** - Tests fetching a specific audiobook by ID
3. **UpdateAudioBook** - Tests updating audiobook information
4. **DeleteAudioBook** - Tests deleting audiobooks
5. **GetPublicAudioBooks** - Tests fetching public audiobooks (no auth required)
6. **GetPublicAudioBook** - Tests fetching a specific public audiobook

### Test Scenarios

Each handler function is tested with multiple scenarios:

- **Success cases** - Valid requests that should succeed
- **Error cases** - Invalid requests that should return appropriate error responses
- **Authentication cases** - Tests for unauthorized access
- **Validation cases** - Tests for invalid input data

## Running Tests

### Run all tests

```bash
go test ./test/... -v
```

### Run specific test file

```bash
go test ./test/audiobooks_test.go -v
```

### Run specific test function

```bash
go test ./test/... -v -run TestGetAudioBooks
```

### Run tests with coverage

```bash
go test ./test/... -v -cover
```

### Run tests with coverage report

```bash
go test ./test/... -v -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Test Dependencies

The tests use the following packages:

- `github.com/stretchr/testify/assert` - For assertions
- `github.com/stretchr/testify/mock` - For mocking
- `github.com/gofiber/fiber/v2` - For HTTP testing
- `net/http/httptest` - For HTTP request testing

## Mock Implementation

The tests use a `MockRepository` that implements the `database.Repository` interface. This allows testing the handlers without requiring a real database connection.

### Mock Setup

```go
mockRepo := new(MockRepository)
mockRepo.On("GetAudioBooksByUser", mock.Anything, userID, 20, 0).Return(audiobooks, 1, nil)
```

### Mock Verification

```go
mockRepo.AssertExpectations(t)
```

## Test Helper Functions

The test file includes several helper functions:

- `createTestHandler()` - Creates a handler with mocked dependencies
- `createTestApp()` - Creates a Fiber app for testing
- `createTestUserContext()` - Creates a test user context
- `createTestAudioBook()` - Creates a test audiobook
- `createTestAudioBookWithDetails()` - Creates a test audiobook with details

## Example Test Structure

```go
func TestGetAudioBooks(t *testing.T) {
    tests := []struct {
        name           string
        setupMock      func(*MockRepository)
        expectedStatus int
        expectedData   bool
    }{
        {
            name: "successful get audiobooks",
            setupMock: func(mockRepo *MockRepository) {
                // Setup mock expectations
            },
            expectedStatus: http.StatusOK,
            expectedData:   true,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Best Practices

1. **Use table-driven tests** - Makes it easy to add new test cases
2. **Mock external dependencies** - Ensures tests are fast and reliable
3. **Test both success and error cases** - Ensures robust error handling
4. **Use descriptive test names** - Makes it clear what each test validates
5. **Verify mock expectations** - Ensures the correct methods are called
6. **Test HTTP status codes** - Ensures proper HTTP responses
7. **Test response structure** - Ensures correct JSON response format

## Adding New Tests

To add tests for new handlers:

1. Create a new test function following the naming convention `TestFunctionName`
2. Use table-driven tests with multiple scenarios
3. Mock the necessary repository methods
4. Test both success and error cases
5. Verify HTTP status codes and response structure
6. Add the test to the appropriate test file or create a new one

## Continuous Integration

These tests can be integrated into CI/CD pipelines to ensure code quality and prevent regressions. The tests are designed to run quickly and provide clear feedback on any failures.
