# Comprehensive API Test Suite

This document provides an overview of the comprehensive test suite created for all API routes defined in `routes.go`.

## Test Files Overview

### 1. `auth_test.go` - Authentication Routes

Tests for `/auth/*` routes:

- **ValidateToken** (`POST /auth/validate`)

  - Missing authorization header
  - Invalid authorization header format
  - Invalid bearer token
  - Valid token scenarios (requires Supabase setup)

- **Me** (`GET /auth/me`)

  - Successful user profile retrieval
  - Unauthorized access

- **HealthCheck** (`GET /auth/health`)
  - Basic health check response

### 2. `cart_test.go` - Cart and Purchase Routes

Tests for `/user/*` cart and purchase routes:

- **AddToCart** (`POST /user/cart`)

  - Successful add to cart
  - Missing audiobook ID
  - Invalid audiobook ID format
  - Database errors

- **RemoveFromCart** (`DELETE /user/cart/:audiobookId`)

  - Successful removal
  - Invalid audiobook ID
  - Database errors

- **GetCart** (`GET /user/cart`)

  - Successful cart retrieval
  - Empty cart
  - Database errors

- **IsInCart** (`GET /user/cart/:audiobookId/check`)

  - Item in cart
  - Item not in cart
  - Invalid audiobook ID
  - Database errors

- **Checkout** (`POST /user/checkout`)

  - Successful checkout
  - Database errors

- **GetPurchaseHistory** (`GET /user/purchases`)

  - Successful history retrieval
  - Database errors

- **IsAudioBookPurchased** (`GET /user/audiobooks/:audiobookId/purchased`)
  - Purchased audiobook
  - Not purchased audiobook
  - Invalid audiobook ID
  - Database errors

### 3. `admin_test.go` - Admin Routes

Tests for `/admin/*` routes:

- **CreateAudioBook** (`POST /admin/audiobooks`)

  - Successful creation from upload
  - Missing required fields
  - Database errors

- **UpdateAudioBookPrice** (`PUT /admin/audiobooks/:id/price`)

  - Successful price update
  - Invalid audiobook ID
  - Audiobook not found
  - Invalid price
  - Database errors

- **CreateUpload** (`POST /admin/uploads`)

  - Successful upload session creation
  - Missing required fields
  - Invalid upload type
  - Database errors

- **GetUploadProgress** (`GET /admin/uploads/:id/progress`)

  - Successful progress retrieval
  - Invalid upload ID
  - Upload not found
  - Access denied scenarios

- **GetUploadDetails** (`GET /admin/uploads/:id`)

  - Successful details retrieval
  - Invalid upload ID
  - Upload not found

- **DeleteUpload** (`DELETE /admin/uploads/:id`)

  - Successful deletion
  - Invalid upload ID
  - Upload not found
  - Access denied scenarios

- **GetJobStatus** (`GET /admin/audiobooks/:id/jobs`)

  - Successful job status retrieval
  - Invalid audiobook ID
  - Database errors

- **TriggerSummarizeAndTagJobs** (`POST /admin/audiobooks/:id/trigger-summarize-tag`)
  - Successful job triggering
  - Invalid audiobook ID
  - Audiobook not found

### 4. `jobs_test.go` - Job Management Routes

Tests for job-related admin routes:

- **UpdateJobStatus** (`POST /admin/jobs/:job_id/status`)

  - Successful status update
  - Invalid job ID
  - Job not found
  - Invalid status
  - Database errors

- **RetryJob** (`POST /admin/audiobooks/:id/jobs/:job_id/retry`)

  - Successful retry
  - Invalid IDs
  - Job not found
  - Job/audiobook mismatch
  - Job not in failed status

- **RetryAllFailedJobs** (`POST /admin/audiobooks/:id/retry-all`)
  - Successful retry of all failed jobs
  - Invalid audiobook ID
  - No failed jobs
  - Database errors

### 5. `internal_test.go` - Internal API Routes

Tests for `/internal/*` routes (API key protected):

- **InternalTriggerSummarizeAndTagJobs** (`POST /internal/audiobooks/:id/trigger-summarize-tag`)

  - Successful internal job triggering
  - Invalid audiobook ID
  - Audiobook not found
  - Audiobook not ready for processing
  - Database errors

- **InternalUpdateJobStatus** (`POST /internal/jobs/:job_id/status`)

  - Successful internal status update
  - Invalid job ID
  - Job not found
  - Invalid status transition
  - Job status with error result
  - Database errors

- **InternalAPIKeyMiddleware** (Conceptual test)
  - Missing API key
  - Invalid API key
  - Valid API key

### 6. `audiobooks_test.go` - Extended Audiobook Routes

Updated existing tests to include:

- All user audiobook operations
- Public audiobook access
- Admin audiobook management
- Complete mock repository implementation

## Mock Implementation

### MockRepository

Comprehensive mock implementation of `database.Repository` interface including:

- All audiobook operations
- Cart and purchase operations
- Upload and file operations
- Processing job operations
- All supporting methods (transcripts, AI outputs, tags, embeddings, etc.)

### Test Helpers

- `createTestHandler()` - Creates handler with mocked dependencies
- `createTestApp()` - Creates Fiber app for testing
- `createTestUserContext()` - Creates test user context
- `createTestAudioBook()` - Creates test audiobook
- `createTestPurchase()` - Creates test purchase record

## Test Coverage

The test suite provides comprehensive coverage for:

### Route Groups

- ✅ Authentication routes (`/auth/*`)
- ✅ Protected user routes (`/user/*`)
- ✅ Admin routes (`/admin/*`)
- ✅ Internal routes (`/internal/*`)

### HTTP Methods

- ✅ GET requests
- ✅ POST requests
- ✅ PUT requests
- ✅ DELETE requests

### Test Scenarios

- ✅ Success cases
- ✅ Validation errors
- ✅ Authentication/authorization errors
- ✅ Database errors
- ✅ Not found errors
- ✅ Access denied errors
- ✅ Invalid input errors

### Error Handling

- ✅ Invalid UUIDs
- ✅ Missing required fields
- ✅ Database connection errors
- ✅ Business logic errors
- ✅ Permission errors

## Running Tests

### Run All Tests

```bash
go test ./test/... -v
```

### Run Specific Test Files

```bash
go test ./test -v -run TestAuth
go test ./test -v -run TestCart
go test ./test -v -run TestAdmin
go test ./test -v -run TestJobs
go test ./test -v -run TestInternal
```

### Run with Coverage

```bash
go test ./test/... -v -cover
go test ./test/... -v -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Dependencies

- `github.com/stretchr/testify/assert` - Assertions
- `github.com/stretchr/testify/mock` - Mocking
- `github.com/gofiber/fiber/v2` - HTTP testing
- `net/http/httptest` - HTTP request testing
- `github.com/google/uuid` - UUID generation

## Notes

1. **Authentication**: Most tests mock authentication by setting user context directly
2. **Database**: All database operations are mocked using testify/mock
3. **External Services**: Supabase storage and Redis queue services are mocked
4. **Validation**: Tests cover both successful validation and validation failures
5. **Error Scenarios**: Comprehensive error testing ensures robust error handling
6. **Real Integration**: For full integration testing, replace mocks with real services

## Future Enhancements

- Add integration tests with real database
- Add performance/load testing
- Add API documentation testing
- Add security testing
- Add middleware testing in isolation
- Add file upload testing with actual files
